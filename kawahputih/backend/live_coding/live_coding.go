// Package live_coding — KODE LATIHAN UKK (Ujian Kompetensi Keahlian).
//
// File ini berisi 3 modul yang paling sering diminta di live coding:
//
//	1. CRUD penuh          → entity Produk (souvenir) + repository + service + handler + route
//	2. Log aktivitas       → pencatatan aksi pengguna (siapa, aksi apa, kapan, hasil) ke tabel audit_logs
//	3. Transaksi database  → pembelian produk dengan stok, atomik (commit/rollback)
//
// Semua pola di sini SAMA PERSIS dengan pola yang dipakai proyek ini
// (handler → usecase → service → repository), jadi jawabanmu akan terlihat
// konsisten dengan aplikasi aslinya. Lihat memo/live_coding.md untuk
// penjelasan konsep, skrip bicara, dan Q&A penguji.
//
// CATATAN: file ini ikut ter-compile (go build ./...) supaya tidak ada
// kesalahan sintaks saat di-copy ke tempat lain. Ia TIDAK dipasang ke router
// utama; jika ingin dipakai sungguhan, panggil RegisterAll dari routes.go.
package live_coding

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/KawahPutih/backend/internal/domain/audit"
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/KawahPutih/backend/pkg/response"
)

// =====================================================================
// MODUL 1 — CRUD PENUH (entity "Produk": souvenir/merchandise)
// =====================================================================
//
// Alur request pada CRUD (ingat urutan ini saat live coding):
//
//	HTTP request (Gin)
//	   │
//	   ▼
//	Handler  → terima JSON, validasi bentuk request, panggil Service
//	   │
//	   ▼
//	Service  → aturan bisnis (validasi nilai, hitung total, dsb.)
//	   │
//	   ▼
//	Repository → satu-satunya yang menyentuh GORM/database
//	   │
//	   ▼
//	Database (MySQL)
//
// Handler TIDAK berbicara dengan database, Repository TIDAK tahu Gin.
// Inilah "clean architecture" sederhana — sebutkan kalau penguji tanya.

// Produk adalah contoh entity: satu baris tabel `produk`.
type Produk struct {
	// ID memakai UUID char(36), konsisten dengan tabel lain di proyek ini.
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // soft delete: baris tidak benar-benar dihapus
	Name      string         `gorm:"type:varchar(150);not null" json:"name"`
	Price     int64          `gorm:"not null" json:"price"`  // harga satuan (rupiah)
	Stock     int            `gorm:"not null" json:"stock"`  // stok tersisa — dipakai juga di modul transaksi
	IsActive  bool           `gorm:"default:true" json:"is_active"`
}

// TableName memberi tahu GORM nama tabelnya; tanpa ini GORM akan menebak "produk".
func (Produk) TableName() string { return "produk" }

// BeforeCreate mengisi ID otomatis sebelum baris disimpan.
func (p *Produk) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}

// ProdukRequest adalah bentuk JSON yang diterima untuk CREATE.
// Tag validate dipakai validator/v10 — sama seperti domain lain di proyek.
type ProdukRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=150"`
	Price int64  `json:"price" validate:"required,min=1"`
	Stock int    `json:"stock" validate:"required,min=0"`
}

// ProdukUpdateRequest untuk UPDATE — semua field opsional.
type ProdukUpdateRequest struct {
	Name     string `json:"name" validate:"omitempty,min=3,max=150"`
	Price    *int64 `json:"price" validate:"omitempty,min=1"`
	Stock    *int   `json:"stock" validate:"omitempty,min=0"`
	IsActive *bool  `json:"is_active"`
}

// ProdukResponse adalah bentuk yang DIKEMBALIKAN ke klien.
// Pisahkan entity dan DTO agar hash password/kolom internal tidak pernah bocor.
type ProdukResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Price     int64  `json:"price"`
	Stock     int    `json:"stock"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// toProdukResponse satu-satunya tempat konversi entity → DTO (mapper).
func toProdukResponse(p *Produk) ProdukResponse {
	return ProdukResponse{
		ID: p.ID, Name: p.Name, Price: p.Price, Stock: p.Stock,
		IsActive: p.IsActive, CreatedAt: p.CreatedAt.Format(time.RFC3339),
	}
}

// Error sentinel milik domain: dideklarasikan SEKALI, dipakai berkali-kali.
// apperror memetakan Kind → status HTTP (NotFound → 404, Conflict → 409, dst.)
var (
	ErrProdukNotFound = apperror.NotFound("Produk tidak ditemukan.")
	ErrProdukExists   = apperror.Conflict("Nama produk sudah dipakai.")
)

// ProdukRepository interface: service tidak mengimpor GORM, hanya interface ini.
// Manfaatnya: bisa di-mock saat testing, dan service tetap bersih.
type ProdukRepository interface {
	Create(ctx context.Context, p *Produk) error
	FindByID(ctx context.Context, id string) (*Produk, error)
	FindByName(ctx context.Context, name string) (*Produk, error)
	List(ctx context.Context, offset, limit int, search string) ([]Produk, int64, error)
	Update(ctx context.Context, p *Produk) error
	Delete(ctx context.Context, id string) error
}

type produkRepository struct{ db *gorm.DB }

func NewProdukRepository(db *gorm.DB) ProdukRepository { return &produkRepository{db: db} }

// Create menyimpan satu baris. ctx dibawa agar timeout/cancel request ikut ke DB.
func (r *produkRepository) Create(ctx context.Context, p *Produk) error {
	return r.db.WithContext(ctx).Create(p).Error
}

// FindByID mengembalikan (nil, nil) bila tidak ketemu — service yang memutuskan
// errornya. Ini pola yang sama dipakai domain auth & user di proyek ini.
func (r *produkRepository) FindByID(ctx context.Context, id string) (*Produk, error) {
	var p Produk
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FindByName dipakai service untuk menolak nama duplikat (unique constraint
// ditangani di sini secara eksplisit agar pesan errornya enak dibaca).
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

// List dengan pagination + pencarian. Offset/limit dibungkus agar nilai
// aneh (0, negatif, terlalu besar) tidak merusak query.
func (r *produkRepository) List(ctx context.Context, offset, limit int, search string) ([]Produk, int64, error) {
	q := r.db.WithContext(ctx).Model(&Produk{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("name LIKE ?", like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []Produk
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// Update menyimpan perubahan field yang terisi saja (partial update).
// Save(p) juga bisa, tapi Updates(map) lebih eksplisit dan anti-lupa-null.
func (r *produkRepository) Update(ctx context.Context, p *Produk) error {
	return r.db.WithContext(ctx).Model(p).Updates(map[string]interface{}{
		"name": p.Name, "price": p.Price, "stock": p.Stock, "is_active": p.IsActive,
	}).Error
}

// Delete soft delete (karena DeletedAt ada di entity): baris ditandai
// deleted_at, tidak benar-benar dihapus — aman bila salah hapus.
func (r *produkRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Produk{}, "id = ?", id).Error
}

// clampPagination menormalkan page/limit dari query string.
func clampPagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

// ProdukService memuat aturan bisnis CRUD.
// TIDAK ada gin.Context di sini — handler yang menerjemahkan HTTP.
type ProdukService struct {
	repo   ProdukRepository
	auditor AktivitasLogger // modul 2: log aktivitas (opsional, bisa nil)
}

func NewProdukService(repo ProdukRepository, auditor AktivitasLogger) *ProdukService {
	return &ProdukService{repo: repo, auditor: auditor}
}

// createLogged menuliskan log aktivitas untuk aksi apapun di service.
// Kegagalan menulis log TIDAK menggagalkan operasi utama — log itu pelengkap.
func (s *ProdukService) createLogged(ctx context.Context, action string, p *Produk) {
	if s.auditor == nil {
		return
	}
	// ExtractAuditInfo mengambil (user_id, role, ip, user-agent) dari context —
	// diisi oleh middleware audit saat request masuk.
	actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
	s.auditor.LogSuccess(ctx, actorID, actorRole, action, "produk", p.ID,
		map[string]interface{}{"name": p.Name, "price": p.Price, "stock": p.Stock}, ip, ua)
}

// Create: cek nama duplikat → simpan → kembalikan DTO.
func (s *ProdukService) Create(ctx context.Context, req ProdukRequest) (*ProdukResponse, error) {
	existing, err := s.repo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if existing != nil {
		return nil, ErrProdukExists
	}
	p := &Produk{Name: req.Name, Price: req.Price, Stock: req.Stock, IsActive: true}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, apperror.Internal(err)
	}
	s.createLogged(ctx, "produk.create", p) // log aktivitas: aksi CREATE
	resp := toProdukResponse(p)
	return &resp, nil
}

// Get: baca satu; nil dari repository → error 404 dari domain.
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

// List: kembalikan data + total untuk meta pagination.
func (s *ProdukService) List(ctx context.Context, page, limit int, search string) ([]ProdukResponse, int64, error) {
	page, limit = clampPagination(page, limit)
	rows, total, err := s.repo.List(ctx, (page-1)*limit, limit, search)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	out := make([]ProdukResponse, 0, len(rows))
	for i := range rows {
		out = append(out, toProdukResponse(&rows[i]))
	}
	return out, total, nil
}

// Update: ambil dulu (ada?), terapkan field yang dikirim, simpan.
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
		// dup != nil && dup.ID != id artinya nama dipakai produk LAIN.
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
	s.createLogged(ctx, "produk.update", p) // log aktivitas: aksi UPDATE
	resp := toProdukResponse(p)
	return &resp, nil
}

// Delete: cek ada → hapus (soft) → 404 bila tidak ada.
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
	s.createLogged(ctx, "produk.delete", p) // log aktivitas: aksi DELETE
	return nil
}

var validate = validator.New()

// validateStruct menerjemahkan validator/v10 → []response.FieldError agar
// bentuk error validasinya sama persis dengan endpoint lain di proyek.
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

// ProdukHandler: satu-satunya bagian yang tahu tentang Gin.
type ProdukHandler struct {
	svc *ProdukService
}

func NewProdukHandler(svc *ProdukService) *ProdukHandler { return &ProdukHandler{svc: svc} }

// Create = POST /produk
func (h *ProdukHandler) Create(c *gin.Context) {
	var req ProdukRequest
	if err := c.ShouldBindJSON(&req); err != nil { // bentuk JSON salah → 400
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := validateStruct(req); errs != nil { // isi tidak valid → 422
		response.ValidationFailed(c, errs)
		return
	}
	data, err := h.svc.Create(c.Request.Context(), req) // context dibawa ke DB
	if err != nil {
		apperror.WriteHTTP(c, err) // Kind → status HTTP (409, 500, ...)
		return
	}
	response.Created(c, "produk created", data) // 201
}

// List = GET /produk?page=1&limit=10&search=kopi
func (h *ProdukHandler) List(c *gin.Context) {
	page := queryInt(c, "page")
	limit := queryInt(c, "limit")
	rows, total, err := h.svc.List(c.Request.Context(), page, limit, c.Query("search"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "produk retrieved", rows, response.BuildMeta(page, limit, total))
}

// Get = GET /produk/:id
func (h *ProdukHandler) Get(c *gin.Context) {
	data, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "produk retrieved", data)
}

// Update = PUT /produk/:id
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

// Delete = DELETE /produk/:id
func (h *ProdukHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "produk deleted", nil)
}

// queryInt membaca query string dengan nilai default, tanpa panic.
func queryInt(c *gin.Context, key string) int {
	n := 0
	_, _ = fmt.Sscanf(c.Query(key), "%d", &n)
	return n
}

// =====================================================================
// MODUL 2 — LOG AKTIVITAS (audit trail)
// =====================================================================
//
// Pertanyaan penguji umumnya: "bagaimana mencatat siapa melakukan apa?"
// Jawaban 3 lapis di proyek ini:
//
//   1) AuditMiddleware  → memasang info pembawa aksi (user_id, role, ip,
//      user-agent) ke request context SETIAP request yang membawa JWT.
//   2) audit.ExtractAuditInfo(ctx) → mengambil info itu kembali dari context
//      di dalam service (tanpa menyentuh gin.Context).
//   3) audit.Service.LogSuccess/LogFailure → menulis 1 baris ke tabel
//      audit_logs (actor, action, target, status, metadata JSON, ip, ua).
//
// Di modul CRUD di atas, setiap Create/Update/Delete memanggil
// s.createLogged(...) — jadi semua perubahan tercatat otomatis.

// AktivitasLogger interface memutus ketergantungan langsung ke *audit.Service.
// *audit.Service memenuhi interface ini secara struktural (duck typing Go).
type AktivitasLogger interface {
	LogSuccess(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string)
}

// AktivitasHandler menyediakan endpoint baca log untuk admin.
// (Menulis log sudah dilakukan di service; membaca log butuh filter.)
type AktivitasHandler struct {
	svc *audit.Service
}

func NewAktivitasHandler(svc *audit.Service) *AktivitasHandler { return &AktivitasHandler{svc: svc} }

// List = GET /admin/aktivitas?actor_id=...&action=produk.create&from=...&to=...
// Semua parameter opsional; kosong = tanpa filter.
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

// =====================================================================
// MODUL 3 — TRANSAKSI DATABASE (pembelian produk, atomik)
// =====================================================================
//
// Kata kunci yang harus diucapkan ke penguji: ATOMICITAS.
// Semua query di dalam db.Transaction(...) dijalankan sebagai satu kesatuan:
//   - return nil   → COMMIT  (semua perubahan disimpan)
//   - return error → ROLLBACK (semua perubahan dibatalkan)
// Bila stok dikurangi tanpa transaksi dan aplikasi mati di tengah jalan,
// stok sudah berkurang tapi pesanan tidak jadi — data jadi tidak konsisten.

// Transaksi adalah bukti pembelian produk (mirip "order").
type Transaksi struct {
	ID           string    `gorm:"type:char(36);primaryKey" json:"id"`
	ProdukID     string    `gorm:"type:char(36);index;not null" json:"produk_id"`
	ProdukName   string    `gorm:"type:varchar(150)" json:"produk_name"`
	BuyerID      string    `gorm:"type:char(36);index" json:"buyer_id"`
	Quantity     int       `gorm:"not null" json:"quantity"`
	PriceAtBuy   int64     `gorm:"not null" json:"price_at_buy"` // harga saat beli (snapshot, bukan harga sekarang)
	Total        int64     `gorm:"not null" json:"total"`
	CreatedAt    time.Time `json:"created_at"`
}

func (Transaksi) TableName() string { return "transaksi" }

func (t *Transaksi) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	return nil
}

var (
	ErrStokHabis  = apperror.Validation("Stok produk tidak mencukupi.")
	ErrQtyInvalid = apperror.Validation("Jumlah pembelian harus lebih dari 0.")
)

// BeliProduk membeli `qty` produk: cek stok + kurangi stok + catat transaksi
// dalam SATU transaksi database. Kalau salah satu gagal, semua dibatalkan.
func BeliProduk(ctx context.Context, db *gorm.DB, produkID string, qty int, buyerID string, auditor AktivitasLogger) (*Transaksi, error) {
	// Validasi awal (sebelum DB) agar query tidak sia-sia.
	if qty <= 0 {
		return nil, ErrQtyInvalid
	}

	var result *Transaksi

	// db.Transaction menjalankan blok ini dalam satu koneksi + satu unit kerja.
	// tx adalah "db versi transaksi": SEMUA query dalam blok wajib lewat tx,
	// bukan db — kalau lewat db, query itu keluar dari transaksi!
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p Produk
		// SELECT ... FOR UPDATE mengunci baris produk sampai transaksi selesai.
		// Tujuannya mencegah "race condition": dua orang membeli item terakhir
		// secara bersamaan — keduanya membaca stok=1, keduanya lolos cek, stok
		// jadi -1. Dengan FOR UPDATE, pembaca kedua MENUNGGU transaksi pertama
		// selesai, lalu membaca stok yang SUDAH dikurangi.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&p, "id = ?", produkID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProdukNotFound // return error → seluruh transaksi ROLLBACK
			}
			return apperror.Internal(err)
		}

		// Cek stok di dalam transaksi (setelah kunci, data dijamin segar).
		if p.Stock < qty {
			return ErrStokHabis // → ROLLBACK, stok tidak berubah
		}

		// Kurangi stok dengan gorm.Expr agar pengurangan terjadi di database
		// (stock - ?), bukan di aplikasi — terhindar dari angka basi.
		if err := tx.Model(&Produk{}).Where("id = ?", produkID).
			Update("stock", gorm.Expr("stock - ?", qty)).Error; err != nil {
			return apperror.Internal(err)
		}

		// Harga di-snapshot saat beli: kalau besok harga naik, transaksi lama
		// tetap menunjukkan harga yang benar-benar dibayar.
		total := p.Price * int64(qty)
		t := &Transaksi{
			ProdukID: p.ID, ProdukName: p.Name, BuyerID: buyerID,
			Quantity: qty, PriceAtBuy: p.Price, Total: total,
		}
		if err := tx.Create(t).Error; err != nil {
			return apperror.Internal(err) // gagal insert → stok dikembalikan otomatis
		}

		result = t // simpan hasil untuk dikembalikan setelah COMMIT
		return nil // nil → COMMIT: stok berkurang + transaksi tersimpan bersamaan
	})

	if err != nil {
		return nil, err // transaksi gagal (rollback) — teruskan error ke handler
	}

	// Log aktivitas DILUAR transaksi: kalau audit gagal, pembelian tetap sah.
	// Sebaliknya jangan taruh di dalam blok tx — log tidak boleh ikut di-rollback.
	if auditor != nil {
		actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
		auditor.LogSuccess(ctx, actorID, actorRole, "transaksi.create", "produk", produkID,
			map[string]interface{}{"qty": qty, "total": result.Total, "produk": result.ProdukName},
			ip, ua)
	}
	return result, nil
}

// TransaksiHandler membungkus BeliProduk untuk endpoint POST /transaksi.
type TransaksiHandler struct {
	db      *gorm.DB
	auditor AktivitasLogger
}

func NewTransaksiHandler(db *gorm.DB, auditor AktivitasLogger) *TransaksiHandler {
	return &TransaksiHandler{db: db, auditor: auditor}
}

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
	// Pembeli adalah user yang sedang login (diisi middleware RequireAuth).
	buyerID := middleware.CurrentUserID(c)
	data, err := BeliProduk(c.Request.Context(), h.db, req.ProdukID, req.Quantity, buyerID, h.auditor)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.Created(c, "transaksi created", data)
}

// Riwayat = GET /transaksi/mine — pembeli melihat transaksinya sendiri.
func (h *TransaksiHandler) Riwayat(c *gin.Context) {
	page := queryInt(c, "page")
	limit := queryInt(c, "limit")
	page, limit = clampPagination(page, limit)

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

// =====================================================================
// PEMASANGAN (wiring) — cara semua modul di atas disambung ke router.
// =====================================================================
//
// RegisterAll dipanggil dari routes.go. Pola pemasangannya:
//
//	1. Buat repository (butuh *gorm.DB)
//	2. Buat service     (butuh repository)
//	3. Buat handler     (butuh service)
//	4. Daftarkan route  (butuh handler + jwtManager untuk middleware auth)
//
// DB produk & transaksi perlu AutoMigrate sekali saat bootstrap (lihat .MD).
func RegisterAll(rg *gin.RouterGroup, db *gorm.DB, jwtManager *jwtutil.Manager, auditor *audit.Service) {
	produkRepo := NewProdukRepository(db)
	produkSvc := NewProdukService(produkRepo, auditor)
	produkH := NewProdukHandler(produkSvc)

	transaksiH := NewTransaksiHandler(db, auditor)
	aktivitasH := NewAktivitasHandler(auditor)

	requireAuth := middleware.RequireAuth(jwtManager)                 // wajib punya JWT valid
	requireAdmin := middleware.RequireRole("ADMIN", "super_admin")   // cek role di claim JWT

	// Publik: siapa pun boleh MELIHAT produk.
	rg.GET("/produk", produkH.List)
	rg.GET("/produk/:id", produkH.Get)

	// Admin: hanya ADMIN/super_admin boleh MENGUBAH produk.
	admin := rg.Group("", requireAuth, requireAdmin)
	admin.POST("/produk", produkH.Create)
	admin.PUT("/produk/:id", produkH.Update)
	admin.DELETE("/produk/:id", produkH.Delete)

	// Transaksi: hanya user yang SUDAH LOGIN (role apa pun) boleh membeli.
	authenticated := rg.Group("", requireAuth)
	authenticated.POST("/transaksi", transaksiH.Create)
	authenticated.GET("/transaksi/mine", transaksiH.Riwayat)

	// Log aktivitas: hanya admin yang boleh membaca jejak audit.
	auditGroup := rg.Group("/admin/aktivitas", requireAuth, requireAdmin)
	auditGroup.GET("", aktivitasH.List)
}
