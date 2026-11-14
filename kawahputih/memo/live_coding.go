// ╔══════════════════════════════════════════════════════════════════════╗
// ║  live_coding.go — KODE LATIHAN UKK (Ujian Kompetensi Keahlian)        ║
// ║  3 modul: CRUD • LOG AKTIVITAS • TRANSAKSI DATABASE                   ║
// ╚══════════════════════════════════════════════════════════════════════╝
//
// KETERANGAN TANDA yang dipakai di seluruh file ini:
//
//   ▶ STEP n        → urutan mengerjakan saat live coding (ikuti berurutan!)
//   💡 PENJELASAN   → penjelasan "kenapa ditulis seperti ini" untuk blok di atasnya
//   ⚠️ PENTING      → jebakan umum / hal yang WAJIB diingat (sering ditanya penguji)
//   🗣️ KATA PENGUJI → kalimat yang bisa langsung diucapkan saat menjelaskan
//   ❌ SALAH / ✅ BENAR → perbandingan cara yang keliru vs yang benar
//   ❓ PENGUJI      → pertanyaan yang sering keluar + ➡️ JAWABAN siap pakai
//   🔗 KAITAN       → hubungan bagian ini dengan bagian lain / aplikasi asli
//
// Kode di file ini polanya SAMA PERSIS dengan aplikasi asli
// (handler → service → repository). Salinan yang ikut ter-compile ada di
// backend/live_coding/live_coding.go. Panduan lengkap: memo/live_coding.md.
package live_coding

// ▶ STEP 0 — IMPORTS
// ─────────────────────────
// Semua import yang dibutuhkan. Jangan hafal satu per satu; hafal KATEGORINYA:
//   • gin          = kerangka HTTP (router, handler)
//   • validator    = validasi struct dengan tag `validate:"..."`
//   • uuid         = pembuat ID unik
//   • gorm + clause = ORM database + helper (kunci baris, dll.)
//   • internal/... = potongan aplikasi sendiri (audit, middleware, error, response)
import (
	"context"  // 💡 dibawa ke repository → kalau user batal request, query ikut dibatalkan
	"errors"   // 💡 untuk errors.Is(...) — cara memeriksa error spesifik (ErrRecordNotFound)
	"fmt"      // 💡 dipakai util kecil queryInt
	"time"     // 💡 tipe waktu untuk kolom created_at/updated_at

	"github.com/gin-gonic/gin"                  // 🔗 framework HTTP yang dipakai seluruh proyek
	"github.com/go-playground/validator/v10"    // 🔗 validator tag `validate:"required,min=3"`
	"github.com/google/uuid"                    // 🔗 ID char(36) — sama seperti tabel lain di proyek
	"gorm.io/gorm"                              // 🔗 ORM: query dibangun dari method chain
	"gorm.io/gorm/clause"                       // 🔗 dipakai modul 3: clause.Locking (FOR UPDATE)

	"github.com/KawahPutih/backend/internal/domain/audit"  // 🔗 log aktivitas
	"github.com/KawahPutih/backend/internal/middleware"    // 🔗 RequireAuth/RequireRole/CurrentUserID
	"github.com/KawahPutih/backend/pkg/apperror"           // 🔗 error domain → status HTTP otomatis
	"github.com/KawahPutih/backend/pkg/jwtutil"            // 🔗 manager JWT untuk middleware auth
	"github.com/KawahPutih/backend/pkg/response"           // 🔗 pembungkus response JSON standar
)

// ══════════════════════════════════════════════════════════════════════
// 📖 KAMUS SINTAKS GO — penjelasan per bentuk sintaks yang dipakai file ini.
// Setiap entri menunjuk ke tempat kemunculannya (STEP) supaya bisa dibaca
// sambil melihat kode aslinya di bawah.
// ══════════════════════════════════════════════════════════════════════
//
// ── 1. package & import ─────────────────────────────────────────────
//   package live_coding
//   → Mendeklarasikan "kode di file ini milik package bernama live_coding".
//     Semua file dalam folder yang sama wajib satu package.
//
//   import "context"            → satu import satu baris.
//   import (
//       "a"
//       "b"
//   )                           → banyak import digroup (standar Go).
//   Urutan: stdlib dulu, kosong satu baris, lalu library pihak ketiga/konco,
//   lalu package internal proyek — inilah konvensi gofmt yang dipakai proyek.
//
// ── 2. deklarasi variabel: var, :=, dan var (...) ───────────────────
//   var total int64            → deklarasi dengan TIPE eksplisit.
//   total := 0                 → deklarasi + inisialisasi; TIPE DISIMPULKAN
//                                dari sisi kanan. `:=` hanya boleh di dalam
//                                fungsi. Di luar fungsi wajib `var`.
//   x = 5                      → `=` tanpa `:` = MEMPERBARUI variabel yang
//                                sudah ada, bukan membuat baru.
//   var (
//       ErrProdukNotFound = ...
//       ErrProdukExists   = ...
//   )                          → merangkai beberapa var dalam satu blok
//                                (dipakai STEP 2 untuk daftar error domain).
//   _, _ = fmt.Sscanf(...)     → `_` = "blank identifier": kotak sampah.
//                                Sscanf mengembalikan 2 nilai, kita tidak
//                                peduli keduanya, jadi dibuang ke `_`.
//   for i := range rows        → juga pakai `_`: `for i := range` saat index
//     out = append(out, ...      tidak dibutuhkan... tapi di STEP 4 kita
//                                memang pakai index `i` karena butuh alamat
//                                elemen: &rows[i].
//
// ── 3. func: parameter, MULTI RETURN, receiver ──────────────────────
//   func clampPagination(page, limit int) (int, int)
//        └─ page DAN limit dua-duanya int (tipe ditulis SEKALI di belakang);
//           mengembalikan DUA nilai (Go punya multi-return bawaan!)
//
//   func (s *ProdukService) Create(ctx context.Context, req ProdukRequest) (*ProdukResponse, error)
//        └──┬──────────┘ └─┬──┘ └────┬─────────────┘ └──────────┬──────────┘
//           │              │         │                          │
//       RECEIVER        nama        parameter                nilai balik
//       (method        fungsi                                (pointer DTO +
//       milik struct)                                        error — pola
//                                                            "data, err")
//   • Receiver `(s *ProdukService)` = Go's "method": fungsi yang menempel
//     pada sebuah tipe. `s` = nama variabel penerima (mirip `this`/`self`
//     di bahasa lain, tapi eksplisit).
//   • Kombinasi (data, err) adalah IDIOM Go: fungsi mengembalikan hasil DAN
//     error sekaligus. Pemanggil WAJIB cek error dulu:
//         data, err := h.svc.Get(...)
//         if err != nil { ... }        ← SELALU sebelum memakai `data`.
//
// ── 4. POINTER: * dan & (paling sering ditanya!) ────────────────────
//   Tanda * dan & punya DUA makna tergantung posisi:
//
//   Tipe:   *Produk          → "pointer ke Produk" (alamat memori struct).
//           *int64           → pointer ke int64 — dipakai DTO update (STEP 1b)
//                              agar bisa membedakan "tidak dikirim" (nil)
//                              dari "dikirim nol" (0).
//   Aksi:   &Produk{}        → `&` = AMBIL ALAMAT: membuat struct lalu
//                              mengoper alamatnya (hemat copy, dan mengizinkan
//                              fungsi mengUBAH aslinya).
//           &rows[i]         → alamat elemen slice ke-i.
//           *req.Price       → `*` di depan variabel = "AMBIL ISI": dereference,
//                              menyalin nilai int64 di balik pointer tersebut.
//   Ringkas: `*` pada TIPE = deklarasi pointer; `&` = ambil alamat;
//            `*` pada VARIABEL = baca isi pointer.
//
//   nil  → nilai nol untuk pointer/interface: "belum menunjuk apa-apa".
//          repository mengembalikan (nil, nil) = "data tidak ada" (STEP 3);
//          if p == nil { return ErrProdukNotFound } (STEP 4).
//
// ── 5. struct + TAG ─────────────────────────────────────────────────
//   type Produk struct {
//       ID    string `gorm:"type:char(36);primaryKey" json:"id"`
//   }
//   → `type X struct { ... }` mendefinisikan tipe baru berisi kolom-kolom.
//   → Bagian dalam backtick `` `...` `` = STRUCT TAG: metadata yang dibaca
//     library lain, BUKAN Go sendiri:
//       gorm:"..."  → dibaca GORM saat migrasi/query:
//                     `type:char(36);primaryKey` — beberapa opsi dipisah `;`
//       json:"..."  → dibaca encoding/json saat marshalling:
//                     nama field di JSON, `json:"-"` = sembunyikan.
//       validate:"required,min=3" → dibaca validator/v10:
//                     beberapa aturan dipisah koma `,`.
//   ⚠️ Tag ditulis dengan BACKTICK (`` ` ``), bukan kutip biasa — dan tidak
//      boleh ada spasi setelah backtick pembuka.
//
// ── 6. interface ────────────────────────────────────────────────────
//   type ProdukRepository interface {
//       Create(ctx context.Context, p *Produk) error
//       ...
//   }
//   → Interface = daftar KONTRAK method. Tipe apapun yang punya SEMUA method
//     itu otomatis memenuhi interface — tanpa `implements` (structural typing).
//     Inilah kenapa *audit.Service bisa langsung jadi AktivitasLogger.
//   → Interface juga bisa jadi "tipe nilai": `var p *Produk = nil` dan
//     `repo ProdukRepository` — service menyimpan apa pun yang memenuhi kontrak.
//
// ── 7. pola pengecekan error: if dengan statement awal ──────────────
//   if err := c.ShouldBindJSON(&req); err != nil {
//        └──────┬──────────────┘  └──┬──────┘
//     STATEMENT AWAL          KONDISI
//   → Go mengizinkan satu statement kecil sebelum `;` di dalam `if`.
//     err hanya hidup DI DALAM blok if — rapi, tidak menodai luar.
//   → idiome paling sering di seluruh kodebase; hafalkan bentuk ini.
//
//   errors.Is(err, gorm.ErrRecordNotFound)
//   → membandingkan error terhadap sentinel yang dibungkus berlapis
//     (kalau error dibungkus fmt.Errorf("%w", ...), == biasa gagal,
//     errors.Is tetap menemukannya).
//
// ── 8. slice: make, append, dan range ───────────────────────────────
//   out := make([]ProdukResponse, 0, len(rows))
//        └─ membuat slice kosong, panjang 0, KAPASITAS awal len(rows)
//           → append tidak perlu memperbesar memori berulang-ulang.
//   out = append(out, toProdukResponse(&rows[i]))
//        → append mengembalikan slice BARU, WAJIB ditampung kembali (`out =`).
//
// ── 9. map[string]interface{} ───────────────────────────────────────
//   map[string]interface{}{"name": p.Name, "price": p.Price}
//   → map dengan kunci string dan nilai "apa saja" (interface{}).
//     Dipakai untuk metadata log & partial update GORM:
//     Updates(map[string]interface{}{...}) → hanya kolom yang ada di map
//     yang di-UPDATE (partial update).
//
// ── 10. anonymous function / closure (dipakai modul 3!) ─────────────
//   db.Transaction(func(tx *gorm.DB) error {
//       ... // tx hanya hidup di sini
//       return nil    // nil → COMMIT
//       return err    // error → ROLLBACK
//   })
//   → Ini FUNGSI TANPA NAMA yang dikirim sebagai ARGUMEN ke Transaction.
//     GORM memanggilnya di tengah: setelah BEGIN, lalu memutuskan
//     COMMIT/ROLLBACK dari nilai baliknya. Pattern "callback" ini membuat
//     aturan commit/rollback jadi otomatis — tidak ada BEGIN/COMMIT manual.
//
// ── 11. method chain GORM (query builder) ───────────────────────────
//   r.db.WithContext(ctx).Model(&Produk{}).Where("id = ?", id).First(&p).Error
//        └────┬───────┘ └────┬────────┘ └─────┬──────────┘ └──┬─────┘ └─┬─┘
//          mulai dari db  pilih tabel       filter          eksekusi   ambil
//          + konteks                                        1 baris    errornya
//   → Setiap method mengembalikan *gorm.DB yang SUDAH berisi kondisi baru,
//     sehingga bisa dirantai. `.Error` = hasil eksekusi (nil = sukses).
//   → "?" dalam Where = placeholder parameter binding (anti SQL injection).
//
// ── 12. fmt.Sscanf ─────────────────────────────────────────────────
//   fmt.Sscanf(c.Query(key), "%d", &n)
//   → mem-parse string jadi angka menurut format "%d" (decimal),
//     hasil disimpan ke ALAMAT variabel n (makanya pakai &n).
//     Gagal parse → n tetap 0 → ditormalkan lagi oleh clampPagination.

// ══════════════════════════════════════════════════════════════════════
// GAMBARAN BESAR — ingat peta ini sebelum menulis apa pun:
//
//   Request HTTP (browser/postman)
//        │
//        ▼
//   MIDDLEWARE  → cek JWT (siapa dia?), cek role (boleh dia?), pasang info audit
//        │
//        ▼
//   HANDLER     → terima JSON, validasi BENTUK, panggil service, bentuk response
//        │
//        ▼
//   SERVICE     → ATURAN BISNIS: cek duplikat, cek stok, hitung total, catat log
//        │
//        ▼
//   REPOSITORY  → satu-satunya yang menyentuh GORM/database
//        │
//        ▼
//   MySQL (tabel produk, transaksi, audit_logs)
//
// 🗣️ KATA PENGUJI:
// "Setiap domain dipisah jadi handler, service, dan repository.
//  Handler tidak menyentuh database, repository tidak tahu tentang Gin.
//  Dengan ini kalau aturan bisnis berubah saya cukup ubah service,
//  dan kalau ganti database saya cukup ubah repository."
// ══════════════════════════════════════════════════════════════════════

// ╔══════════════════════════════════════════════════════════════════╗
// ║ MODUL 1 — CRUD PENUH (entity "Produk": souvenir/merchandise)      ║
// ║                                                                   ║
// ║ Urutan pengerjaan yang aman (lakukan STEP 1→6 berurutan):         ║
// ║   STEP 1: entity + DTO      (bentuk data — paling mudah duluan)   ║
// ║   STEP 2: error sentinel    (daftar error khas domain)            ║
// ║   STEP 3: repository        (semua query database)                ║
// ║   STEP 4: service           (aturan bisnis)                       ║
// ║   STEP 5: handler           (HTTP: bind, validasi, response)      ║
// ║   STEP 6: route             (sambungkan + middleware)             ║
// ╚══════════════════════════════════════════════════════════════════╝

// ▶ STEP 1a — ENTITY: satu struct = satu tabel database
// ────────────────────────────────────────────────────
type Produk struct {
	// ID pakai UUID char(36), konsisten dengan tabel lain (users, bookings, dll.)
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // ← kehadiran field ini mengaktifkan SOFT DELETE
	Name      string         `gorm:"type:varchar(150);not null" json:"name"`
	Price     int64          `gorm:"not null" json:"price"` // harga satuan rupiah (int64 → tidak ada pecahan rupiah)
	Stock     int            `gorm:"not null" json:"stock"` // stok — dipakai juga modul 3 (transaksi)
	IsActive  bool           `gorm:"default:true" json:"is_active"`
}

func (Produk) TableName() string { return "produk" }

func (p *Produk) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}

// 💡 PENJELASAN:
// • "GORM tags" (gorm:"...") mengatur SKEMA DB; tag `json:"..."` mengatur
//   BENTUK API. Satu struct melayani dua dunia — tapi nanti DTO tetap dipisah.
// • `json:"-"` pada DeletedAt = jangan pernah kirim kolom ini ke klien.
// • int64 untuk harga: rupiah tidak punya sen, jadi integer cukup dan bebas
//   dari bug pembulatan float.
//
// ⚠️ PENTING:
// • Tanpa `TableName()`, GORM menebak nama tabel dari nama struct
//   ("produk" kebetulan cocok) — tulis eksplisit agar tidak bergantung tebakan.
// • Soft delete: baris tidak dihapus fisik, hanya `deleted_at` diisi waktu.
//   Semua query GORM otomatis mengabaikan baris yang sudah di-soft-delete.
//
// ❓ PENGUJI: "Kenapa ID pakai UUID, bukan auto increment?"
// ➡️ JAWABAN: "UUID tidak bisa ditebak (aman dari enumerasi URL) dan bisa
//    dibuat di aplikasi tanpa round-trip ke database untuk minta ID dulu."
//
// ❓ PENGUJI: "Apa itu soft delete?"
// ➡️ JAWABAN: "Data ditandai waktu terhapus, bukan dihapus. Bisa dipulihkan,
//    dan riwayat transaksi yang menunjuk ke data itu tidak rusak."

// ▶ STEP 1b — DTO: bentuk data yang MELEWATI API (bukan entity mentah)
// ─────────────────────────────────────────────────────────────────────
type ProdukRequest struct { // ← bentuk yang DITERIMA saat CREATE/POST
	Name  string `json:"name" validate:"required,min=3,max=150"`
	Price int64  `json:"price" validate:"required,min=1"`
	Stock int    `json:"stock" validate:"required,min=0"`
}

type ProdukUpdateRequest struct { // ← bentuk yang DITERIMA saat UPDATE/PUT
	Name     string `json:"name" validate:"omitempty,min=3,max=150"` // omitempty = boleh dikosongkan
	Price    *int64 `json:"price" validate:"omitempty,min=1"`        // pointer: bisa BEDAKAN "tidak dikirim" vs "dikirim 0"
	Stock    *int   `json:"stock" validate:"omitempty,min=0"`
	IsActive *bool  `json:"is_active"`
}

type ProdukResponse struct { // ← bentuk yang DIKIRIM kembali ke klien
	ID        string `json:"id"`
	Name      string `json:"name"`
	Price     int64  `json:"price"`
	Stock     int    `json:"stock"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// toProdukResponse = MAPPER: satu-satunya tempat konversi entity → DTO.
func toProdukResponse(p *Produk) ProdukResponse {
	return ProdukResponse{
		ID: p.ID, Name: p.Name, Price: p.Price, Stock: p.Stock,
		IsActive: p.IsActive, CreatedAt: p.CreatedAt.Format(time.RFC3339),
	}
}

// 💡 PENJELASAN:
// • Kenapa DTO dipisah dari entity? Entity punya kolom internal (misal
//   PasswordHash di domain auth). Kalau entity langsung jadi response,
//   kolom rahasia itu bisa bocor. DTO = gerbang penyaring data keluar.
// • Field pointer (*int64, *bool) di update request adalah trik PARTIAL
//   UPDATE: JSON {"price":0} vs tanpa "price" bisa dibedakan — dengan int64
//   biasa, nilai kosong dan nol akan tertukar.
// • Tag `validate:"..."` dibaca oleh validator/v10 di STEP 5 (handler).
//
// 🗣️ KATA PENGUJI: "Saya memisahkan entity dan DTO supaya kolom internal
// seperti password hash tidak pernah bocor ke response API, dan supaya
// kontrak API tidak berubah setiap kali skema database berubah."

// ▶ STEP 2 — ERROR SENTINEL: daftar error khas domain ini
// ────────────────────────────────────────────────────────
var (
	ErrProdukNotFound = apperror.NotFound("Produk tidak ditemukan.")   // → HTTP 404
	ErrProdukExists   = apperror.Conflict("Nama produk sudah dipakai.") // → HTTP 409
)

// 💡 PENJELASAN:
// • "Sentinel" = error yang dibuat SEKALI dan dibandingkan pakai == atau
//   errors.As. Semua tempat yang butuh "produk tidak ada" memakai objek
//   yang sama — konsisten, tidak ada dua pesan berbeda untuk keadaan sama.
// • apperror punya taksonomi Kind: Validation→422, NotFound→404,
//   Conflict→409, Unauthorized→401, Forbidden→403, Internal→500.
//   Handler cukup memanggil apperror.WriteHTTP(c, err) — pemetaan otomatis.
//
// ❌ SALAH:  return errors.New("not found")  → handler tidak tahu itu 404,
//            semuanya jadi 500.
// ✅ BENAR:  return ErrProdukNotFound        → 404 otomatis lewat apperror.
//
// 🔗 KAITAN: domain lain di proyek (auth, booking) punya daftar serupa
//    di file errors.go — kebiasaan yang sama, tampilkan saat ditanya.

// ▶ STEP 3 — REPOSITORY: satu-satunya lapisan yang menyentuh database
// ────────────────────────────────────────────────────────────────────
// Interface dulu: service hanya berbicara dengan interface ini, BUKAN dengan
// gorm.DB. Manfaatnya: (1) service bisa dites tanpa database (pakai mock),
// (2) ganti database cukup ganti implementasi repository.
type ProdukRepository interface {
	Create(ctx context.Context, p *Produk) error
	FindByID(ctx context.Context, id string) (*Produk, error)
	FindByName(ctx context.Context, name string) (*Produk, error)
	List(ctx context.Context, offset, limit int, search string) ([]Produk, int64, error)
	Update(ctx context.Context, p *Produk) error
	Delete(ctx context.Context, id string) error
}

// Implementasi struct (di bawah interface) — ini yang benar-benar memegang gorm.
type produkRepository struct{ db *gorm.DB }

// Constructor: pola New* dipakai di seluruh proyek.
func NewProdukRepository(db *gorm.DB) ProdukRepository { return &produkRepository{db: db} }

func (r *produkRepository) Create(ctx context.Context, p *Produk) error {
	// WithContext(ctx): kalau request dibatalkan/timeout, query ikut berhenti.
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *produkRepository) FindByID(ctx context.Context, id string) (*Produk, error) {
	var p Produk
	// First(...) = ambil 1 baris + error kalau tidak ada (ErrRecordNotFound).
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // ← KONVENSI proyek ini: "tidak ada" = (nil, nil), BUKAN error
	}
	if err != nil {
		return nil, err // error teknis (koneksi putus, dsb.) — diteruskan
	}
	return &p, nil
}

func (r *produkRepository) FindByName(ctx context.Context, name string) (*Produk, error) {
	var p Produk
	err := r.db.WithContext(ctx).First(&p, "name = ?", name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *produkRepository) List(ctx context.Context, offset, limit int, search string) ([]Produk, int64, error) {
	// Model(&Produk{}) = titik awal query; chain method seperti query builder.
	q := r.db.WithContext(ctx).Model(&Produk{})
	if search != "" {
		like := "%" + search + "%"
		// "?" adalah PARAMETER BINDING — nilai dikirim terpisah dari SQL,
		// sehingga isi search tidak pernah dieksekusi sebagai SQL (anti-injection).
		q = q.Where("name LIKE ?", like)
	}
	// Count dulu untuk total (dipakai pagination), baru ambil halamannya.
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []Produk
	// Order DESC = terbaru di atas; Offset/Limit = potongan halaman.
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// Updates(map) = PARTIAL UPDATE: hanya kolom yang ada di map yang berubah.
// (Alternatif Save(p) menimpa SEMUA kolom — berbahaya kalau field lupa diisi.)
func (r *produkRepository) Update(ctx context.Context, p *Produk) error {
	return r.db.WithContext(ctx).Model(p).Updates(map[string]interface{}{
		"name": p.Name, "price": p.Price, "stock": p.Stock, "is_active": p.IsActive,
	}).Error
}

// Delete soft delete (karena entity punya DeletedAt): GORM otomatis
// mengubah jadi UPDATE deleted_at = NOW(), bukan DELETE FROM.
func (r *produkRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Produk{}, "id = ?", id).Error
}

// clampPagination menormalkan input page/limit dari query string.
func clampPagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // batas atas: orang jahat tidak bisa minta 1.000.000 baris
	}
	return page, limit
}

// 💡 PENJELASAN:
// • Konvensi "tidak ditemukan = (nil, nil)" membuat service yang memutuskan
//   errornya: ErrProdukNotFound di STEP 4. Repository tidak punya opini.
// • Setiap Where pakai "?" — jelaskan ke penguji: "inilah yang mencegah
//   SQL injection; nilai dikirim sebagai parameter, bukan ditempel ke string SQL."
//
// ⚠️ PENTING (3 jebakan favorit penguji):
//   1. Lupa WithContext(ctx) → query jalan terus walau user sudah tutup halaman.
//   2. string SQL disusun manual: Where("name LIKE '%" + search + "%'")
//      ❌ SALAH — pintu SQL injection. ✅ BENAR: parameter binding dengan "?".
//   3. Count dan Find harus pakai query `q` yang SAMA supaya filter pencarian
//      ikut di kedua-duanya (kalau tidak, total dan data tidak sinkron).

// ▶ STEP 4 — SERVICE: aturan bisnis. TIDAK ada gin.Context di sini!
// ────────────────────────────────────────────────────────────────────
type ProdukService struct {
	repo    ProdukRepository // bukan *gorm.DB — service melihat interface saja
	auditor AktivitasLogger  // modul 2: log aktivitas (opsional, boleh nil)
}

func NewProdukService(repo ProdukRepository, auditor AktivitasLogger) *ProdukService {
	return &ProdukService{repo: repo, auditor: auditor}
}

// createLogged: pembungkus kecil supaya setiap aksi CRUD tercatat (modul 2).
// Kegagalan menulis log TIDAK menggagalkan operasi utama — log itu pelengkap.
func (s *ProdukService) createLogged(ctx context.Context, action string, p *Produk) {
	if s.auditor == nil {
		return // auditor opsional: kalau tidak dipasang, lewati saja
	}
	// ExtractAuditInfo mengambil (user_id, role, ip, user-agent) dari context —
	// diisi oleh middleware audit saat request masuk (modul 2).
	actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
	s.auditor.LogSuccess(ctx, actorID, actorRole, action, "produk", p.ID,
		map[string]interface{}{"name": p.Name, "price": p.Price, "stock": p.Stock}, ip, ua)
}

// CREATE: cek nama duplikat → simpan → catat log → kembalikan DTO.
func (s *ProdukService) Create(ctx context.Context, req ProdukRequest) (*ProdukResponse, error) {
	existing, err := s.repo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, apperror.Internal(err) // error teknis dibungkus → 500 rapi
	}
	if existing != nil {
		return nil, ErrProdukExists // aturan bisnis: nama harus unik → 409
	}
	p := &Produk{Name: req.Name, Price: req.Price, Stock: req.Stock, IsActive: true}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, apperror.Internal(err)
	}
	s.createLogged(ctx, "produk.create", p) // 🔎 log aktivitas: CREATE
	resp := toProdukResponse(p)
	return &resp, nil
}

// GET SATU: repository bilang nil → service mengubahnya jadi 404.
func (s *ProdukService) Get(ctx context.Context, id string) (*ProdukResponse, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if p == nil {
		return nil, ErrProdukNotFound
	}
	resp := toProdukResponse(p)
	return &resp, nil
}

// LIST: kembalikan data + total (total dipakai untuk meta pagination).
func (s *ProdukService) List(ctx context.Context, page, limit int, search string) ([]ProdukResponse, int64, error) {
	page, limit = clampPagination(page, limit)
	rows, total, err := s.repo.List(ctx, (page-1)*limit, limit, search) // offset = (halaman-1) × jumlah-per-halaman
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	out := make([]ProdukResponse, 0, len(rows))
	for i := range rows {
		out = append(out, toProdukResponse(&rows[i]))
	}
	return out, total, nil
}

// UPDATE: ambil dulu (ada?) → terapkan field yang DIKIRIM saja → simpan.
func (s *ProdukService) Update(ctx context.Context, id string, req ProdukUpdateRequest) (*ProdukResponse, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if p == nil {
		return nil, ErrProdukNotFound
	}
	if req.Name != "" {
		dup, err := s.repo.FindByName(ctx, req.Name)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		// Kunci logika: dup != nil && dup.ID != id → nama dipakai produk LAIN.
		// Kalau dup.ID == id, itu produk sendiri → tidak apa-apa (nama sama).
		if dup != nil && dup.ID != id {
			return nil, ErrProdukExists
		}
		p.Name = req.Name
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.Stock != nil {
		p.Stock = *req.Stock
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.Internal(err)
	}
	s.createLogged(ctx, "produk.update", p) // 🔎 log aktivitas: UPDATE
	resp := toProdukResponse(p)
	return &resp, nil
}

// DELETE: cek ada → hapus (soft) → selesai.
func (s *ProdukService) Delete(ctx context.Context, id string) error {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return apperror.Internal(err)
	}
	if p == nil {
		return ErrProdukNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperror.Internal(err)
	}
	s.createLogged(ctx, "produk.delete", p) // 🔎 log aktivitas: DELETE
	return nil
}

// 💡 PENJELASAN:
// • Pattern cek-ubah-simpan pada Update: FindByID dulu supaya 404 muncul
//   SEBELUM kita repot menyiapkan perubahan, dan supaya kita tahu ID lama.
// • createLogged dipanggil SETELAH operasi sukses — log sukses hanya untuk
//   aksi yang benar-benar terjadi. (Kalau gagal, jalur error tidak sampai sini.)
//
// ⚠️ PENTING: JANGAN memanggil gin.Context di service. Kenapa?
//   Supaya service bisa dipanggil dari mana saja (cron, worker, test).
//   Handler-lah yang menerjemahkan HTTP ↔ domain.
//
// ❓ PENGUJI: "Kenapa validasi dilakukan dua tempat?"
// ➡️ JAWABAN: "Handler memvalidasi BENTUK request (field kosong, panjang).
//    Service menjaga ATURAN BISNIS (nama duplikat, stok cukup), karena
//    service bisa dipanggil dari sumber lain, bukan hanya HTTP."

// ▶ STEP 5 — HANDLER: satu-satunya lapisan yang tahu tentang Gin
// ───────────────────────────────────────────────────────────────
// Handler HANYA melakukan 3 hal (hafalkan ritme ini):
//   1) bind JSON        → c.ShouldBindJSON
//   2) validasi bentuk  → validateStruct
//   3) panggil service  → lalu response.OK / apperror.WriteHTTP
type ProdukHandler struct {
	svc *ProdukService
}

func NewProdukHandler(svc *ProdukService) *ProdukHandler { return &ProdukHandler{svc: svc} }

// CREATE = POST /produk → 201 Created
func (h *ProdukHandler) Create(c *gin.Context) {
	var req ProdukRequest
	if err := c.ShouldBindJSON(&req); err != nil { // JSON rusak/bukan JSON → 400
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := validateStruct(req); errs != nil { // isi tidak valid → 422
		response.ValidationFailed(c, errs)
		return
	}
	data, err := h.svc.Create(c.Request.Context(), req) // context dibawa ke DB
	if err != nil {
		apperror.WriteHTTP(c, err) // Kind error → status HTTP yang benar otomatis
		return
	}
	response.Created(c, "produk created", data) // 201 + envelope standar
}

// LIST = GET /produk?page=1&limit=10&search=kopi
func (h *ProdukHandler) List(c *gin.Context) {
	page := queryInt(c, "page")
	limit := queryInt(c, "limit")
	rows, total, err := h.svc.List(c.Request.Context(), page, limit, c.Query("search"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	// OKWithMeta = data + meta {page, limit, total, total_pages} — bentuk
	// yang sama dipakai semua list endpoint proyek (frontend mengandalkannya).
	response.OKWithMeta(c, "produk retrieved", rows, response.BuildMeta(page, limit, total))
}

// GET = GET /produk/:id
func (h *ProdukHandler) Get(c *gin.Context) {
	data, err := h.svc.Get(c.Request.Context(), c.Param("id")) // c.Param = path param
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "produk retrieved", data)
}

// UPDATE = PUT /produk/:id
func (h *ProdukHandler) Update(c *gin.Context) {
	var req ProdukUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := validateStruct(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	data, err := h.svc.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "produk updated", data)
}

// DELETE = DELETE /produk/:id
func (h *ProdukHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "produk deleted", nil)
}

// queryInt: baca query string jadi angka, nilai salah = 0 (service yang menormalkan).
func queryInt(c *gin.Context, key string) int {
	n := 0
	_, _ = fmt.Sscanf(c.Query(key), "%d", &n)
	return n
}

// validateStruct: validator/v10 → []response.FieldError, bentuk error yang
// SAMA dengan semua endpoint lain (frontend membaca field+message ini).

// Validator instance dibuat SEKALI global (mahal kalau dibuat per request).
var validate = validator.New()

func validateStruct(s interface{}) []response.FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var out []response.FieldError
	for _, fe := range err.(validator.ValidationErrors) {
		out = append(out, response.FieldError{Field: fe.Field(), Message: fe.Field() + " tidak valid"})
	}
	return out
}

// 💡 PENJELASAN:
// • "Envelope standar" response proyek: { success, message, data, meta? }.
//   Frontend mem-parsing bentuk ini di semua halaman — jadi endpoint baru
//   WAJIB mengikuti bentuk yang sama (response.OK/Created/OKWithMeta).
// • urutan check: bind (400) → validate (422) → service (404/409/500).
//   Jangan tukar urutan — validasi bentuk lebih murah daripada query DB.
//
// ❓ PENGUJI: "Kenapa error ditangani apperror.WriteHTTP, bukan c.JSON manual?"
// ➡️ JAWABAN: "Supaya semua endpoint memetakan error dengan cara yang sama.
//    Domain cuma menyebut jenis masalah (NotFound/Conflict/...), lapisan
//    response yang menentukan kode HTTP dan format JSON-nya."

// ▶ STEP 6 — ROUTE: sambungkan handler ke URL + middleware (lihat RegisterAll
//             di bagian bawah file ini, bagian PEMASANGAN).

// ╔══════════════════════════════════════════════════════════════════╗
// ║ MODUL 2 — LOG AKTIVITAS (audit trail)                             ║
// ║                                                                   ║
// ║ Tujuan: menjawab "SIAPA melakukan APA pada OBJEK apa, KAPAN,      ║
// ║ dan hasilnya SUKSES atau GAGAL?" — untuk jejak & investigasi.     ║
// ╚══════════════════════════════════════════════════════════════════╝
//
// ALUR 3 LAPIS (hafalkan, ini jawaban utama modul ini):
//
//   LAPIS 1: AuditMiddleware      → SETIAP request yang bawa JWT:
//             pasang user_id/role/ip/user-agent ke request context.
//   LAPIS 2: audit.ExtractAuditInfo(ctx) → di DALAM service, ambil kembali
//             info itu dari context (tanpa menyentuh gin.Context).
//   LAPIS 3: LogSuccess/LogFailure → tulis 1 baris ke tabel audit_logs.
//
// 🔗 KAITAN: di modul CRUD, setiap Create/Update/Delete memanggil
//    s.createLogged(...) — jadi SEMUA perubahan data tercatat otomatis.

// AktivitasLogger = interface agar service TIDAK bergantung langsung ke
// *audit.Service (aturan: domain tidak saling mengimpor — dihubungkan via
// interface yang dipasang di bootstrap). *audit.Service memenuhi interface
// ini secara struktural tanpa perlu tahu satu sama lain.
type AktivitasLogger interface {
	LogSuccess(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string)
}

// AktivitasHandler: endpoint BACA log (menulis log sudah dilakukan service).
type AktivitasHandler struct {
	svc *audit.Service
}

func NewAktivitasHandler(svc *audit.Service) *AktivitasHandler { return &AktivitasHandler{svc: svc} }

// LIST = GET /admin/aktivitas
// Filter opsional: actor_id, action, target_type, target_id, status, from, to.
// Contoh: /admin/aktivitas?action=produk.create&from=2026-11-01
func (h *AktivitasHandler) List(c *gin.Context) {
	page := queryInt(c, "page")
	limit := queryInt(c, "limit")
	rows, total, err := h.svc.List(
		c.Request.Context(),
		page, limit,
		c.Query("actor_id"), c.Query("action"),
		c.Query("target_type"), c.Query("target_id"),
		c.Query("status"), c.Query("from"), c.Query("to"),
	)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "aktivitas retrieved", rows, response.BuildMeta(page, limit, total))
}

// Contoh ISI SATU BARIS tabel audit_logs setelah admin membuat produk:
//
// {
//   "actor_id":   "b1a2...uuid",           ← siapa (dari JWT)
//   "actor_role": "ADMIN",                  ← role saat aksi
//   "action":     "produk.create",          ← apa (konvensi: entitas.aksi)
//   "target_type":"produk",
//   "target_id":  "9f8e...uuid",            ← objek yang dikenai aksi
//   "status":     "success",                ← success / failure
//   "metadata":   {"name":"Kopi Ciwidey","price":25000,"stock":10}, ← detail (JSON)
//   "ip_address": "127.0.0.1",
//   "user_agent": "Mozilla/5.0 ..."
// }
//
// ⚠️ PENTING (2 prinsip yang sering ditanya):
//   1. Log SUKSES ditulis SETELAH operasi sukses; log GAGAL (LogFailure)
//      ditulis di jalur error — percobaan gagal (mis. login salah) justru
//      sering paling penting untuk dianalisis.
//   2. Gagal menulis log TIDAK BOLEH menggagalkan operasi utama.
//      Log itu pelengkap, bukan penghalang.
//
// ❓ PENGUJI: "Bedanya log aktivitas dengan log aplikasi (zap)?"
// ➡️ JAWABAN: "Log aplikasi untuk DEVELOPER (teknis, hilang saat restart,
//    susah difilter per user). Log aktivitas untuk BISNIS/AUDIT: tersimpan
//    di database, terstruktur, bisa difilter per pelaku/aksi/waktu lewat API."

// ╔══════════════════════════════════════════════════════════════════╗
// ║ MODUL 3 — TRANSAKSI DATABASE (pembelian produk, atomik)           ║
// ║                                                                   ║
// ║ DUA kata kunci yang wajib diucapkan:                             ║
// ║   1) ATOMICITY  — semua gagal ATAU semua berhasil                 ║
// ║   2) RACE CONDITION — dua pembeli bersamaan pada stok terakhir    ║
// ╚══════════════════════════════════════════════════════════════════╝

// Transaksi = bukti pembelian produk (mirip "order").
type Transaksi struct {
	ID         string    `gorm:"type:char(36);primaryKey" json:"id"`
	ProdukID   string    `gorm:"type:char(36);index;not null" json:"produk_id"`
	ProdukName string    `gorm:"type:varchar(150)" json:"produk_name"`
	BuyerID    string    `gorm:"type:char(36);index" json:"buyer_id"` // siapa pembeli (dari JWT)
	Quantity   int       `gorm:"not null" json:"quantity"`
	PriceAtBuy int64     `gorm:"not null" json:"price_at_buy"` // SNAPSHOT harga saat beli
	Total      int64     `gorm:"not null" json:"total"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Transaksi) TableName() string { return "transaksi" }

func (t *Transaksi) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	return nil
}

var (
	ErrStokHabis  = apperror.Validation("Stok produk tidak mencukupi.") // → 422
	ErrQtyInvalid = apperror.Validation("Jumlah pembelian harus lebih dari 0.")
)

// BeliProduk: cek stok + kurangi stok + catat transaksi dalam SATU transaksi DB.
// ⚠️ Semua query dalam blok wajib lewat `tx`, BUKAN `db` — query lewat db
//    keluar dari transaksi dan tidak ikut rollback!
func BeliProduk(ctx context.Context, db *gorm.DB, produkID string, qty int, buyerID string, auditor AktivitasLogger) (*Transaksi, error) {
	// Validasi murah dulu (sebelum menyentuh DB).
	if qty <= 0 {
		return nil, ErrQtyInvalid
	}

	var result *Transaksi

	// ── TRANSAKSI MULAI ────────────────────────────────────────────────
	// Aturan besar:
	//   return nil   → COMMIT   (semua perubahan DISIMPAN bersamaan)
	//   return error → ROLLBACK (semua perubahan DIBATALKAN bersamaan)
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p Produk

		// (1) KUNCI BARIS: SELECT ... FOR UPDATE.
		// Dua pembeli bersamaan pada item terakhir: keduanya membaca stok=1,
		// keduanya lolos cek, stok jadi -1 (race condition). Dengan FOR UPDATE,
		// pembaca kedua MENUNGGU sampai transaksi pertama selesai, lalu membaca
		// stok yang SUDAH dikurangi — tidak mungkin oversell.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&p, "id = ?", produkID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProdukNotFound // → ROLLBACK otomatis
			}
			return apperror.Internal(err)
		}

		// (2) CEK STOK — di DALAM transaksi (setelah kunci, data dijamin segar).
		if p.Stock < qty {
			return ErrStokHabis // → ROLLBACK: stok tidak berubah
		}

		// (3) KURANGI STOK — lewat gorm.Expr agar pengurangan dieksekusi di
		// DATABASE ("stock - ?"), bukan hitung di aplikasi dengan angka basi.
		if err := tx.Model(&Produk{}).Where("id = ?", produkID).
			Update("stock", gorm.Expr("stock - ?", qty)).Error; err != nil {
			return apperror.Internal(err)
		}

		// (4) CATAT TRANSAKSI — harga di-SNAPSHOT: besok kalau harga naik,
		// transaksi lama tetap menunjukkan harga yang benar-benar dibayar.
		total := p.Price * int64(qty)
		t := &Transaksi{
			ProdukID: p.ID, ProdukName: p.Name, BuyerID: buyerID,
			Quantity: qty, PriceAtBuy: p.Price, Total: total,
		}
		if err := tx.Create(t).Error; err != nil {
			return apperror.Internal(err) // gagal insert → stok otomatis kembali
		}

		result = t // simpan untuk dikembalikan SETELAH commit sukses
		return nil // ← nil = COMMIT: stok berkurang + transaksi tersimpan bersamaan
	})
	// ── TRANSAKSI SELESAI ──────────────────────────────────────────────

	if err != nil {
		return nil, err // rollback sudah terjadi otomatis; teruskan error
	}

	// Log aktivitas ditulis DI LUAR transaksi (SETelah commit):
	// • Pembelian sah tetap tercatat walau penulisan log gagal.
	// • Kalau log ditulis di dalam blok tx, log bisa IKUT di-rollback —
	//   padahal data yang di-rollback itu justru yang perlu diinvestigasi!
	if auditor != nil {
		actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
		auditor.LogSuccess(ctx, actorID, actorRole, "transaksi.create", "produk", produkID,
			map[string]interface{}{"qty": qty, "total": result.Total, "produk": result.ProdukName},
			ip, ua)
	}
	return result, nil
}

// TransaksiHandler: endpoint pembelian.
type TransaksiHandler struct {
	db      *gorm.DB
	auditor AktivitasLogger
}

func NewTransaksiHandler(db *gorm.DB, auditor AktivitasLogger) *TransaksiHandler {
	return &TransaksiHandler{db: db, auditor: auditor}
}

// CREATE = POST /transaksi
func (h *TransaksiHandler) Create(c *gin.Context) {
	var req struct {
		ProdukID string `json:"produk_id" validate:"required"`
		Quantity int    `json:"quantity" validate:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := validateStruct(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	// Pembeli = user yang sedang login (JWT dipasang middleware RequireAuth).
	buyerID := middleware.CurrentUserID(c)
	data, err := BeliProduk(c.Request.Context(), h.db, req.ProdukID, req.Quantity, buyerID, h.auditor)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.Created(c, "transaksi created", data)
}

// RIWAYAT = GET /transaksi/mine — pembeli hanya melihat transaksinya SENDIRI.
func (h *TransaksiHandler) Riwayat(c *gin.Context) {
	page := queryInt(c, "page")
	limit := queryInt(c, "limit")
	page, limit = clampPagination(page, limit)

	// buyerID diambil dari TOKEN, bukan dari parameter request!
	// Ini prinsip keamanan: user tidak bisa melihat riwayat orang lain
	// dengan mengubah ID di URL.
	buyerID := middleware.CurrentUserID(c)

	var rows []Transaksi
	var total int64
	q := h.db.WithContext(c.Request.Context()).Model(&Transaksi{}).Where("buyer_id = ?", buyerID)
	if err := q.Count(&total).Error; err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	if err := q.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&rows).Error; err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	response.OKWithMeta(c, "transaksi retrieved", rows, response.BuildMeta(page, limit, total))
}

// ❓ PENGUJI: "Kapan COMMIT, kapan ROLLBACK?"
// ➡️ JAWABAN: "Function transaksi return nil → COMMIT. Return error →
//    ROLLBACK otomatis. Jadi cukup 'return ErrStokHabis' untuk membatalkan
//    seluruh perubahan — tidak perlu rollback manual."
//
// ❓ PENGUJI: "Kenapa pakai FOR UPDATE?"
// ➡️ JAWABAN: "Untuk mencegah race condition pada stok. Tanpa kunci baris,
//    dua request bersamaan membaca stok yang sama sebelum ada yang mengubah."
//
// ❓ PENGUJI: "Kalau aplikasi mati di tengah transaksi?"
// ➡️ JAWABAN: "Transaksi belum di-commit → database membatalkan sendiri
//    (rollback otomatis oleh database saat koneksi putus). Tidak ada data
//    setengah jadi."
//
// ❌ SALAH: mengurangi stok dulu (db langsung), baru insert transaksi —
//           kalau insert gagal, stok sudah hilang → data tidak konsisten.
// ✅ BENAR: kedua operasi di dalam SATU blok Transaction → atomik.

// ╔══════════════════════════════════════════════════════════════════╗
// ║ PEMASANGAN (WIRING) — menyambung semuanya ke router               ║
// ╚══════════════════════════════════════════════════════════════════╝
//
// Pola pemasangan 4 langkah (sama untuk SEMUA fitur di proyek ini):
//   1. Buat repository (butuh *gorm.DB)
//   2. Buat service    (butuh repository + logger)
//   3. Buat handler    (butuh service)
//   4. Daftarkan route (butuh handler + jwtManager + middleware)
//
// Tabel baru (produk, transaksi) perlu AutoMigrate sekali saat bootstrap —
// lihat panduan langkah lengkap di memo/live_coding.md bagian 5.
func RegisterAll(rg *gin.RouterGroup, db *gorm.DB, jwtManager *jwtutil.Manager, auditor *audit.Service) {
	// 1→2→3: rangkai lapisan.
	produkRepo := NewProdukRepository(db)
	produkSvc := NewProdukService(produkRepo, auditor) // auditor = log aktivitas (modul 2)
	produkH := NewProdukHandler(produkSvc)

	transaksiH := NewTransaksiHandler(db, auditor)
	aktivitasH := NewAktivitasHandler(auditor)

	// 4: middleware.
	requireAuth := middleware.RequireAuth(jwtManager)               // wajib bawa JWT valid
	requireAdmin := middleware.RequireRole("ADMIN", "super_admin") // cek role di claim JWT

	// ── Publik: siapa pun boleh MELIHAT (prinsip least privilege) ──
	rg.GET("/produk", produkH.List)
	rg.GET("/produk/:id", produkH.Get)

	// ── Admin: hanya ADMIN/super_admin boleh MENGUBAH ──
	admin := rg.Group("", requireAuth, requireAdmin)
	admin.POST("/produk", produkH.Create)
	admin.PUT("/produk/:id", produkH.Update)
	admin.DELETE("/produk/:id", produkH.Delete)

	// ── Transaksi: harus LOGIN (role apa pun) boleh membeli ──
	authenticated := rg.Group("", requireAuth)
	authenticated.POST("/transaksi", transaksiH.Create)
	authenticated.GET("/transaksi/mine", transaksiH.Riwayat)

	// ── Log aktivitas: hanya admin boleh MEMBACA jejak audit ──
	auditGroup := rg.Group("/admin/aktivitas", requireAuth, requireAdmin)
	auditGroup.GET("", aktivitasH.List)
}

// 💡 PENJELASAN TABEL ROUTE (sebutkan saat presentasi):
//
//   Method | Path                 | Middleware          | Kenapa
//   GET    | /produk, /produk/:id | —                   | baca = publik
//   POST   | /produk              | auth + ADMIN        | ubah data = admin
//   PUT    | /produk/:id          | auth + ADMIN        | idem
//   DELETE | /produk/:id          | auth + ADMIN        | idem
//   POST   | /transaksi           | auth                | pembeli = user login
//   GET    | /transaksi/mine      | auth                | riwayat milik sendiri
//   GET    | /admin/aktivitas     | auth + ADMIN        | jejak audit sensitif
//
// ⚠️ PENTING: urutan middleware = urutan argumen Group: auth dulu (siapa dia?),
//    baru role (boleh dia?). Jangan dibalik — cek role tanpa tahu siapa
//    usernya tidak berarti.
//
// 🗣️ KATA PENGUJI (penutup yang bagus):
// "Dengan struktur ini, menambah fitur baru hanya mengikuti pola yang sama:
//  entity → repository → service → handler → route. Semua endpoint punya
//  bentuk response, penanganan error, dan log aktivitas yang konsisten."
