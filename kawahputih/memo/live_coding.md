# Live Coding UKK — Panduan & Penjelasan

> Kode lengkapnya ada di **`backend/live_coding/live_coding.go`** (ter-compile, bukan pseudo-code).
> Penjelasan per baris sudah ditulis sebagai komentar `//` di bawah setiap blok kode di file itu.
> File ini berisi **konsep, skrip bicara untuk penguji, dan Q&A**.

---

## 1. Peta Arsitektur (hafalkan ini dulu)

Semua fitur di proyek ini mengikuti alur yang sama:

```
HTTP Request (Gin router)
      │
      ▼
Middleware (auth JWT, role, rate limit, audit info)
      │
      ▼
Handler      ← menerima JSON, memvalidasi BENTUK request, memanggil service
      │
      ▼
Service      ← aturan BISNIS (hitung total, cek stok, cek duplikat, log aktivitas)
      │
      ▼
Repository   ← satu-satunya yang menyentuh database (GORM)
      │
      ▼
Database (MySQL)
```

**Kalimat kunci untuk penguji:**
> "Setiap domain dipisah menjadi handler, service, dan repository.
> Handler tidak menyentuh database, repository tidak tahu tentang Gin.
> Ini membuat kode mudah diuji dan diubah tanpa menyentuh bagian lain."

---

## 2. Modul 1 — CRUD

### 2.1 Apa yang diminta penguji biasanya
"Tambahkan fitur X (misal: produk/souvenir/event) lengkap dengan create, read,
update, delete, sementara pengunjung hanya boleh melihat."

### 2.2 Lima langkah yang ditulis di `live_coding.go`

| Langkah | Bagian kode | Isinya |
|---|---|---|
| 1 | `type Produk struct` + `TableName()` + `BeforeCreate()` | Entity = 1 tabel DB |
| 2 | `ProdukRequest` / `ProdukResponse` | DTO: bentuk masuk & keluar API |
| 3 | `ProdukRepository` + `produkRepository` | Semua query GORM |
| 4 | `ProdukService` | Aturan bisnis: cek duplikat, 404, log aktivitas |
| 5 | `ProdukHandler` + `RegisterAll()` | Binding JSON → response + route |

### 2.3 Poin bicara per bagian

- **Entity**: "Struktur ini memetakan satu baris tabel `produk`. UUID dipakai
  sebagai primary key, konsisten dengan tabel lain di aplikasi.
  `DeletedAt` mengaktifkan *soft delete* — baris tidak benar-benar dihapus."
- **DTO**: "Saya memisahkan entity dan DTO supaya kolom internal (misal
  `password_hash`) tidak pernah bocor ke response API."
- **Repository**: "Repository mengembalikan `(nil, nil)` kalau data tidak
  ditemukan — service yang memutuskan error 404-nya. Query selalu dibawa
  `context` agar timeout request ikut membatalkan query."
- **Service**: "Aturan bisnis ada di sini: nama duplikat ditolak dengan
  `409 Conflict`, data tidak ada → `404`. Semua error dibungkus `apperror`
  yang otomatis dipetakan ke status HTTP."
- **Handler**: "Handler hanya melakukan tiga hal: bind JSON, validasi bentuk,
  dan memetakan error ke response. Tidak ada query SQL di handler."
- **Route**: "Endpoint tulis (POST/PUT/DELETE) saya lindungi middleware
  `RequireAuth` + `RequireRole("ADMIN", "super_admin")`, sedangkan endpoint
  baca terbuka untuk publik. Ini prinsip *least privilege*."

### 2.4 Tabel endpoint CRUD

| Method | Path | Siapa yang boleh | Keterangan |
|---|---|---|---|
| GET | `/produk?page=&limit=&search=` | Publik | List + pagination + filter |
| GET | `/produk/:id` | Publik | Detail |
| POST | `/produk` | ADMIN / super_admin | Create (201) |
| PUT | `/produk/:id` | ADMIN / super_admin | Update parsial |
| DELETE | `/produk/:id` | ADMIN / super_admin | Soft delete |

---

## 3. Modul 2 — Log Aktivitas (audit trail)

### 3.1 Konsep yang harus diucapkan

> "Log aktivitas menjawab empat pertanyaan: **siapa** melakukan **apa**
> terhadap **objek apa**, **kapan**, dan **hasilnya sukses atau gagal**.
> Datanya disimpan di tabel `audit_logs`: actor (user id + role), action
> (mis. `produk.create`), target (tipe + id), status, metadata JSON,
> IP address, dan user-agent."

### 3.2 Alur 3 lapis (sesuai kode di `live_coding.go`)

```
1. AuditMiddleware (per request)  → taruh user_id/role/ip/ua ke request context
2. audit.ExtractAuditInfo(ctx)    → ambil info itu kembali di dalam service
3. LogSuccess / LogFailure        → tulis 1 baris ke tabel audit_logs
```

Di modul CRUD, setiap `Create/Update/Delete` memanggil `s.createLogged(...)`.
Poin penting yang patut disebut:
- "Gagal menulis log **tidak menggagalkan** operasi utama — log itu pelengkap,
  bukan penghalang transaksi."
- "Kegagalan login/ganti password dicatat sebagai `LogFailure` — percobaan
  yang gagal justru sering paling penting untuk dianalisis."
- "Endpoint baca log hanya untuk admin: `GET /admin/aktivitas` dengan filter
  `actor_id`, `action`, `from`, `to`."

### 3.3 Contoh isi satu baris log

```json
{
  "actor_id": "b1a2...uuid", "actor_role": "ADMIN",
  "action": "produk.create", "target_type": "produk", "target_id": "9f8e...",
  "status": "success",
  "metadata": {"name": "Kopi Ciwidey", "price": 25000, "stock": 10},
  "ip_address": "127.0.0.1", "user_agent": "Mozilla/5.0 ..."
}
```

---

## 4. Modul 3 — Transaksi Database (atomik)

### 4.1 Kata kunci: ATOMICITAS + RACE CONDITION

> "Semua query di dalam `db.Transaction(func(tx) error {...})` dijalankan
> sebagai **satu kesatuan**. Return `nil` → **COMMIT** (semua disimpan).
> Return error → **ROLLBACK** (semua dibatalkan). Kalau stok berhasil
> dikurangi tapi insert transaksi gagal, stok otomatis kembali —
> data tidak pernah setengah jadi."

### 4.2 Alur `BeliProduk` (baca ulang komentar di kodenya)

```
BEGIN TRANSACTION
  1. SELECT ... FOR UPDATE  → kunci baris produk (mencegah race condition)
  2. Cek stok ≥ qty          → kurang? return ErrStokHabis → ROLLBACK
  3. UPDATE stock = stock - qty (di database, bukan di aplikasi)
  4. INSERT transaksi (harga di-snapshot saat beli)
  return nil → COMMIT (keduanya berhasil bersama)
```

Poin bicara tambahan:
- **`FOR UPDATE`** (clause.Locking): "Tanpa ini, dua orang yang membeli item
  terakhir secara bersamaan sama-sama membaca stok = 1 dan keduanya lolos —
  stok jadi -1. Dengan kunci baris, pembaca kedua menunggu, lalu membaca stok
  yang sudah berkurang."
- **`gorm.Expr("stock - ?", qty)`**: "Pengurangan dilakukan di database,
  bukan `p.Stock - qty` di aplikasi, supaya tidak memakai angka basi."
- **Snapshot harga** (`price_at_buy`): "Besok kalau harga naik, transaksi lama
  tetap menunjukkan harga yang benar-benar dibayar."
- **Log di luar transaksi**: "Log aktivitas ditulis SETELAH commit — kalau di
  dalam blok tx, log bisa ikut di-rollback padahal pembelian sudah sah."

---

## 5. Jika diminta "pasang ke aplikasi asli"

`bootstrap.App` belum mengekspos `DB` dan `*audit.Service`, jadi butuh 2
langkah kecil sebelum memasang route:

1. **Ekspos DB & audit service** — di `backend/internal/bootstrap/bootstrap.go`:
   tambahkan dua field pada struct `App` dan isinya di `Build()`:
   ```go
   type App struct {
       ...
       DB          *gorm.DB      // ← tambahkan
       AuditService *audit.Service // ← tambahkan (service yang sama yang dipakai AuditHandler)
   }
   // di akhir Build(), sebelum return:
   //   App{ ... DB: db, AuditService: auditSvc, ... }
   ```
2. **Migrasi tabel** — di `cmd/server/bootstrap.go`, tambahkan ke daftar
   AutoMigrate yang sudah ada:
   ```go
   if err := db.AutoMigrate(&live_coding.Produk{}, &live_coding.Transaksi{}); err != nil { ... }
   ```
3. **Daftarkan route** — di `backend/internal/routes/routes.go` (fungsi `NewRouter`):
   ```go
   live_coding.RegisterAll(v1, app.DB, app.JWTManager, app.AuditService)
   ```
4. Restart server → tabel `produk` & `transaksi` dibuat otomatis.

## 6. Cara mendemokan / menguji cepat (curl)

```bash
# Login admin untuk dapat JWT (ganti sesuai akun seed)
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@kawahputih.id","password":"<password>"}' \
  | jq -r .data.access_token)

# CRUD
curl -X POST http://localhost:8080/api/v1/produk -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Kopi Ciwidey","price":25000,"stock":10}'

curl http://localhost:8080/api/v1/produk?search=kopi
curl -X PUT http://localhost:8080/api/v1/produk/<ID> -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" -d '{"price":27000}'
curl -X DELETE http://localhost:8080/api/v1/produk/<ID> -H "Authorization: Bearer $TOKEN"

# Transaksi
curl -X POST http://localhost:8080/api/v1/transaksi -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" -d '{"produk_id":"<ID>","quantity":2}'
curl http://localhost:8080/api/v1/transaksi/mine -H "Authorization: Bearer $TOKEN"

# Log aktivitas (admin)
curl "http://localhost:8080/api/v1/admin/aktivitas?action=produk.create" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 7. Q&A cepat untuk penguji

| Pertanyaan penguji | Jawaban singkat |
|---|---|
| Kenapa pakai handler/service/repository? | Pemisahan tanggung jawab: HTTP vs aturan bisnis vs akses data. Mudah diuji, mudah diganti. |
| Kenapa ID pakai UUID? | Tidak bisa ditebak (aman dari enumerasi), bisa dibuat di aplikasi tanpa round-trip ke DB. |
| Apa itu soft delete? | Baris ditandai `deleted_at`, tidak dihapus fisik. Bisa dipulihkan, query otomatis mengabaikannya. |
| Kenapa validasi dua kali (handler & service)? | Handler memvalidasi *bentuk* request; service menjaga *aturan bisnis* (duplikat, stok) karena bisa dipanggil dari tempat lain. |
| Bagaimana mencegah SQL injection? | GORM selalu pakai parameter binding (`Where("id = ?", id)`), tidak pernah menyusun string SQL manual. |
| Apa itu race condition di stok? | Dua request bersamaan membaca stok sama sebelum diubah. Solusi: `SELECT ... FOR UPDATE` di dalam transaksi. |
| Kapan COMMIT, kapan ROLLBACK? | Function transaksi return `nil` → COMMIT; return error → ROLLBACK otomatis. |
| Kenapa log aktivitas di luar transaksi? | Log tidak boleh ikut di-rollback; pembelian yang sah tetap tercatat walau penulisan log gagal. |
| Bagaimana password disimpan? | Bcrypt (hash + salt otomatis), cost 12 — tidak pernah plaintext. |
| Bagaimana role dicek? | JWT berisi claim role; middleware `RequireRole` membandingkan sebelum handler berjalan. |

---

## 8. Kamus Sintaks Go — penjelasan per bentuk sintaks

> Versi lengkap (12 entri + diseksi baris) ada di header `memo/live_coding.go`
> bagian **📖 KAMUS SINTAKS**. Ringkasan tabel:

| Sintaks | Contoh dari kode | Artinya |
|---|---|---|
| `package x` | `package live_coding` | Deklarasi kepemilikan file ke sebuah package |
| `var total int64` vs `total := 0` | STEP 4 | Deklarasi bertipe vs deklarasi yang tipenya disimpulkan (`:=` hanya di dalam fungsi) |
| `x = 5` | Update field | Memperbarui variabel yang sudah ada (bukan membuat baru) |
| `_, _ = f()` | `queryInt` | Blank identifier — membuang nilai balik yang tidak dipakai |
| `func (s *T) M(...)` | semua method | *Receiver* = method menempel pada struct; `s` mirip `this`, tapi eksplisit |
| `(data, err)` | semua fungsi | Multi-return Go — idiom "hasil + error"; cek `err` dulu sebelum pakai `data` |
| `*T` (pada tipe) | `*ProdukResponse`, `*int64` | Pointer ke tipe; `*int64` di DTO update membedakan "tidak dikirim" (nil) vs "dikirim 0" |
| `&x` | `&Produk{}`, `&rows[i]` | Ambil alamat (mengoper aslinya, bukan salinannya) |
| `*v` (pada variabel) | `*req.Price` | Dereference — baca isi pointer |
| `nil` | `return nil, nil` di repository | "tidak menunjuk apa-apa"; konvensi proyek: data tidak ada = `(nil, nil)` |
| struct tag | `` `gorm:"..." json:"..." validate:"..."` `` | Metadata dibaca library: `;` pemisah opsi gorm, `,` pemisah aturan validate, `json:"-"` = sembunyikan field |
| `interface { M() error }` | `ProdukRepository`, `AktivitasLogger` | Kontrak method; tipe mana pun yang punya semua methodnya otomatis memenuhi (tanpa `implements`) |
| `if err := f(); err != nil {` | semua handler | `if` dengan statement awal — `err` hanya hidup di dalam blok if |
| `errors.Is(err, X)` | STEP 3 | Membandingkan error terhadap sentinel walaupun terbungkus berlapis |
| `make([]T, 0, n)` | `List` | Slice dengan kapasitas awal → `append` tidak boros realokasi |
| `out = append(out, x)` | `List` | `append` mengembalikan slice baru — wajib ditampung kembali |
| `map[string]interface{}` | metadata log, `Updates(map)` | Nilai "apa saja"; dipakai GORM untuk partial update (kolom yang ada di map saja) |
| `func(tx *gorm.DB) error { ... }` | `db.Transaction(...)` | *Anonymous function / closure* sebagai callback: `return nil` → COMMIT, `return error` → ROLLBACK |
| `db.Where("id = ?", id)` | semua repository | Method chain + parameter binding `?` (anti SQL injection); `.Error` = hasil eksekusi |
| `fmt.Sscanf(s, "%d", &n)` | `queryInt` | Parse string → angka, ditulis ke alamat `&n`; gagal parse = 0 |

**Diseksi baris paling sering ditanya:**

```go
func (s *ProdukService) Create(ctx context.Context, req ProdukRequest) (*ProdukResponse, error)
```
`func` + receiver `(s *ProdukService)` → method milik struct; `ctx context.Context` dibawa hingga ke database; balikannya **pasangan** pointer-DTO + error.

```go
if err := c.ShouldBindJSON(&req); err != nil {
```
Statement awal `err := ...` dijalankan dulu, lalu dikondisikan; `&req` mengoper alamat agar hasil bind diisi ke `req` yang sudah kita punya (bukan salinannya).

```go
db.WithContext(ctx).Model(&Produk{}).Where("id = ?", id).First(&p).Error
```
Chain: mulai dari koneksi + context → pilih tabel → filter → eksekusi ambil 1 baris (hasil diisi ke `&p`) → ambil status errornya.

---

## 9. Tips saat live coding

1. **Mulai dari entity + DTO** (paling mudah, menetapkan bentuk data).
2. Kerjakan berurutan: entity → repository → service → handler → route.
   Setiap selesai satu lapis, `go build` — jangan menumpuk error.
3. Salin pola dari **domain yang sudah ada** (mis. `internal/domain/faq`)
   — konsisten lebih dihargai daripada pintar.
4. Kalau kehabisan waktu, utamakan: **Create + Get + route** — lalu ucapkan
   "Update dan Delete polanya sama: FindByID → ubah field → Save".
5. Sebutkan istilah kunci dengan santai: *atomicity, soft delete, least
   privilege, parameter binding, race condition*.
