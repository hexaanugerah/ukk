package booking

import (
	"context"
	"testing"
	"time"

	"github.com/KawahPutih/backend/internal/domain/audit"
	"github.com/KawahPutih/backend/internal/domain/tickettype"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// testAuditLogger adalah implementasi AuditLogger untuk pengujian
type testAuditLogger struct {
	svc *audit.Service
}

func (a *testAuditLogger) LogSuccess(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string) {
	a.svc.LogSuccess(ctx, actorID, actorRole, action, targetType, targetID, metadata, ip, ua)
}

func (a *testAuditLogger) LogFailure(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string) {
	a.svc.LogFailure(ctx, actorID, actorRole, action, targetType, targetID, metadata, ip, ua)
}

// testTicketPricer adalah implementasi TicketPricer untuk pengujian
type testTicketPricer struct {
	svc  *tickettype.Service
	sold Repository
}

func (a *testTicketPricer) GetMany(ctx context.Context, ids []string) ([]TicketInfo, error) {
	types, err := a.svc.GetMany(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]TicketInfo, 0, len(types))
	for _, t := range types {
		out = append(out, TicketInfo{
			ID: t.ID, Name: t.Name, Price: t.Price, Capacity: t.Capacity, IsActive: t.IsActive,
		})
	}
	return out, nil
}

func (a *testTicketPricer) SoldByTicketOn(ctx context.Context, date string) (map[string]int64, error) {
	return a.sold.SoldByTicketOn(ctx, date)
}

// setupTestDB membuat database SQLite in-memory untuk pengujian.
func setupTestDB(t *testing.T) *gorm.DB {
	// Pakai database in-memory unik per tes (tanpa cache=shared)
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate tabel
	err = db.AutoMigrate(
		&Booking{},
		&BookingItem{},
		&tickettype.TicketType{},
		&audit.AuditLog{},
	)
	require.NoError(t, err)

	return db
}

// seedTestTickets membuat jenis tiket uji di database.
func seedTestTickets(t *testing.T, db *gorm.DB) map[string]*tickettype.TicketType {
	tickets := map[string]*tickettype.TicketType{
		"ENTRANCE-UMUM": {
			Code: "ENTRANCE-UMUM", Name: "Tiket Masuk Umum", Category: "ENTRANCE",
			Price: 15000, Capacity: 500, IsActive: true,
		},
		"ENTRANCE-ANAK": {
			Code: "ENTRANCE-ANAK", Name: "Tiket Masuk Anak", Category: "ENTRANCE",
			Price: 10000, Capacity: 500, IsActive: true,
		},
		"TRANSPORT-SHUTTLE": {
			Code: "TRANSPORT-SHUTTLE", Name: "Shuttle Bus PP", Category: "TRANSPORT",
			Price: 25000, Capacity: 400, IsActive: true,
		},
	}

	for _, tt := range tickets {
		tt.CreatedAt = time.Now()
		tt.UpdatedAt = time.Now()
		require.NoError(t, db.Create(tt).Error)
	}
	return tickets
}

func generateTestID() string {
	return "test-" + time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

// TestBookingCreateSuccess menguji pembuatan booking yang berhasil.
func TestBookingCreateSuccess(t *testing.T) {
	db := setupTestDB(t)
	ticketMap := seedTestTickets(t, db)

	bookingRepo := NewRepository(db)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &testAuditLogger{svc: auditService}

	// Mock payment starter
	paymentStarter := &mockPaymentStarter{}

	svc := NewService(bookingRepo, paymentStarter, nil, auditLogger)
	uc := NewUsecase(svc, &testTicketPricer{svc: ticketTypeService, sold: bookingRepo})

	ctx := context.Background()
	userID := "user-123"

	req := CreateRequest{
		VisitDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		Visitor: struct {
			Name  string `json:"name" validate:"required,min=2,max=150"`
			Email string `json:"email" validate:"required,email,max=150"`
			Phone string `json:"phone" validate:"required,min=8,max=20"`
		}{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "08123456789",
		},
		Items: []CreateItem{
			{TicketID: ticketMap["ENTRANCE-UMUM"].ID, Quantity: 2},
			{TicketID: ticketMap["TRANSPORT-SHUTTLE"].ID, Quantity: 1},
		},
	}

	result, err := uc.CreateBooking(ctx, userID, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, StatusPending, result.Booking.Status)
	assert.Equal(t, int64(55000), result.Booking.Subtotal) // 2*15000 + 1*25000
	assert.Equal(t, 3, result.Booking.TotalTickets)
	assert.NotEmpty(t, result.Booking.BookingCode)
	assert.NotNil(t, result.Payment)
}

// TestBookingCreateInvalidDate menguji booking dengan tanggal lampau.
func TestBookingCreateInvalidDate(t *testing.T) {
	db := setupTestDB(t)
	ticketMap := seedTestTickets(t, db)

	bookingRepo := NewRepository(db)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &testAuditLogger{svc: auditService}
	paymentStarter := &mockPaymentStarter{}

	svc := NewService(bookingRepo, paymentStarter, nil, auditLogger)
	uc := NewUsecase(svc, &testTicketPricer{svc: ticketTypeService, sold: bookingRepo})

	ctx := context.Background()
	userID := "user-123"

	req := CreateRequest{
		VisitDate: time.Now().AddDate(0, 0, -1).Format("2006-01-02"), // Yesterday
		Visitor: struct {
			Name  string `json:"name" validate:"required,min=2,max=150"`
			Email string `json:"email" validate:"required,email,max=150"`
			Phone string `json:"phone" validate:"required,min=8,max=20"`
		}{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "08123456789",
		},
		Items: []CreateItem{{TicketID: ticketMap["ENTRANCE-UMUM"].ID, Quantity: 1}},
	}

	_, err := uc.CreateBooking(ctx, userID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Tanggal kunjungan tidak valid.")
}

// TestBookingCreateInvalidQuantity menguji booking dengan jumlah nol.
func TestBookingCreateInvalidQuantity(t *testing.T) {
	db := setupTestDB(t)
	ticketMap := seedTestTickets(t, db)

	bookingRepo := NewRepository(db)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &testAuditLogger{svc: auditService}
	paymentStarter := &mockPaymentStarter{}

	svc := NewService(bookingRepo, paymentStarter, nil, auditLogger)
	uc := NewUsecase(svc, &testTicketPricer{svc: ticketTypeService, sold: bookingRepo})

	ctx := context.Background()
	userID := "user-123"

	req := CreateRequest{
		VisitDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		Visitor: struct {
			Name  string `json:"name" validate:"required,min=2,max=150"`
			Email string `json:"email" validate:"required,email,max=150"`
			Phone string `json:"phone" validate:"required,min=8,max=20"`
		}{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "08123456789",
		},
		Items: []CreateItem{{TicketID: ticketMap["ENTRANCE-UMUM"].ID, Quantity: 0}},
	}

	_, err := uc.CreateBooking(ctx, userID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Jumlah tiket tidak valid.")
}

// TestBookingCreateNegativeQuantity menguji booking dengan jumlah negatif.
func TestBookingCreateNegativeQuantity(t *testing.T) {
	db := setupTestDB(t)
	ticketMap := seedTestTickets(t, db)

	bookingRepo := NewRepository(db)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &testAuditLogger{svc: auditService}
	paymentStarter := &mockPaymentStarter{}

	svc := NewService(bookingRepo, paymentStarter, nil, auditLogger)
	uc := NewUsecase(svc, &testTicketPricer{svc: ticketTypeService, sold: bookingRepo})

	ctx := context.Background()
	userID := "user-123"

	req := CreateRequest{
		VisitDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		Visitor: struct {
			Name  string `json:"name" validate:"required,min=2,max=150"`
			Email string `json:"email" validate:"required,email,max=150"`
			Phone string `json:"phone" validate:"required,min=8,max=20"`
		}{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "08123456789",
		},
		Items: []CreateItem{{TicketID: ticketMap["ENTRANCE-UMUM"].ID, Quantity: -1}},
	}

	_, err := uc.CreateBooking(ctx, userID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Jumlah tiket tidak valid.")
}

// TestBookingCreateNoItems menguji booking tanpa item.
func TestBookingCreateNoItems(t *testing.T) {
	db := setupTestDB(t)
	_ = seedTestTickets(t, db)

	bookingRepo := NewRepository(db)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &testAuditLogger{svc: auditService}
	paymentStarter := &mockPaymentStarter{}

	svc := NewService(bookingRepo, paymentStarter, nil, auditLogger)
	uc := NewUsecase(svc, &testTicketPricer{svc: ticketTypeService, sold: bookingRepo})

	ctx := context.Background()
	userID := "user-123"

	req := CreateRequest{
		VisitDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		Visitor: struct {
			Name  string `json:"name" validate:"required,min=2,max=150"`
			Email string `json:"email" validate:"required,email,max=150"`
			Phone string `json:"phone" validate:"required,min=8,max=20"`
		}{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "08123456789",
		},
		Items: []CreateItem{},
	}

	_, err := uc.CreateBooking(ctx, userID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Mohon pilih minimal satu tiket.")
}

// TestBookingCreateInactiveTicket menguji booking dengan jenis tiket tidak aktif. Penjaga regresi: IsActive harus diturunkan dari Status (NONAKTIF berarti tidak aktif) dan booking harus menolaknya dengan BOOK_009.
func TestBookingCreateInactiveTicket(t *testing.T) {
	db := setupTestDB(t)
	_ = seedTestTickets(t, db)

	bookingRepo := NewRepository(db)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &testAuditLogger{svc: auditService}
	paymentStarter := &mockPaymentStarter{}

	svc := NewService(bookingRepo, paymentStarter, nil, auditLogger)
	uc := NewUsecase(svc, &testTicketPricer{svc: ticketTypeService, sold: bookingRepo})

	ctx := context.Background()

	// Buat tiket tidak aktif lewat service. Inilah perilaku yang dulu salah pada service tickettype (mengabaikan Status), makanya tes ini dulu di-skip.
	inactive, err := ticketTypeService.Create(ctx, "admin-1", tickettype.UpsertRequest{
		Code: "INACTIVE-TKT", Name: "Tiket Nonaktif", Category: "ENTRANCE",
		Price: 15000, Capacity: 100, Status: "NONAKTIF",
	})
	require.NoError(t, err)
	require.NotNil(t, inactive)
	assert.False(t, inactive.IsActive, "Create must persist IsActive=false for status NONAKTIF")

	req := CreateRequest{
		VisitDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		Visitor: struct {
			Name  string `json:"name" validate:"required,min=2,max=150"`
			Email string `json:"email" validate:"required,email,max=150"`
			Phone string `json:"phone" validate:"required,min=8,max=20"`
		}{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "08123456789",
		},
		Items: []CreateItem{{TicketID: inactive.ID, Quantity: 1}},
	}

	_, err = uc.CreateBooking(ctx, "user-123", req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Data tiket tidak ditemukan.")
}

// TestBookingCreateExceedsCapacity menguji booking yang melebihi kapasitas.
func TestBookingCreateExceedsCapacity(t *testing.T) {
	db := setupTestDB(t)
	_ = seedTestTickets(t, db)

	// Buat tiket berkapasitas kecil
	lowCapTicket := &tickettype.TicketType{
		Code: "LOW-CAP", Name: "Low Capacity", Category: "ENTRANCE",
		Price: 10000, Capacity: 2, IsActive: true,
	}
	lowCapTicket.CreatedAt = time.Now()
	lowCapTicket.UpdatedAt = time.Now()
	require.NoError(t, db.Create(lowCapTicket).Error)

	// Jual 2 tiket lebih dulu (penuhi kapasitas)
	existingBooking := &Booking{
		BookingCode:  "KP-20260101-000001",
		UserID:       "other-user",
		VisitorName:  "Other User",
		Email:        "other@example.com",
		Phone:        "08111111111",
		VisitDate:    time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		Status:       StatusPaid,
		Subtotal:     20000,
		ServiceFee:   400,
		Total:        20400,
		TotalTickets: 2,
		Items: []BookingItem{
			{TicketID: lowCapTicket.ID, TicketName: "Low Capacity", Price: 10000, Quantity: 2},
		},
	}
	require.NoError(t, db.Create(existingBooking).Error)

	bookingRepo := NewRepository(db)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &testAuditLogger{svc: auditService}
	paymentStarter := &mockPaymentStarter{}

	svc := NewService(bookingRepo, paymentStarter, nil, auditLogger)
	uc := NewUsecase(svc, &testTicketPricer{svc: ticketTypeService, sold: bookingRepo})

	ctx := context.Background()
	userID := "user-123"

	req := CreateRequest{
		VisitDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		Visitor: struct {
			Name  string `json:"name" validate:"required,min=2,max=150"`
			Email string `json:"email" validate:"required,email,max=150"`
			Phone string `json:"phone" validate:"required,min=8,max=20"`
		}{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "08123456789",
		},
		Items: []CreateItem{{TicketID: lowCapTicket.ID, Quantity: 1}},
	}

	_, err := uc.CreateBooking(ctx, userID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Tiket pada tanggal tersebut tidak tersedia.")
}

// TestBookingCancelSuccess menguji pembatalan booking yang berhasil.
func TestBookingCancelSuccess(t *testing.T) {
	db := setupTestDB(t)
	ticketMap := seedTestTickets(t, db)

	bookingRepo := NewRepository(db)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &testAuditLogger{svc: auditService}
	paymentStarter := &mockPaymentStarter{}

	svc := NewService(bookingRepo, paymentStarter, nil, auditLogger)
	uc := NewUsecase(svc, &testTicketPricer{svc: ticketTypeService, sold: bookingRepo})

	ctx := context.Background()
	userID := "user-123"

	// Buat booking pending dulu
	req := CreateRequest{
		VisitDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		Visitor: struct {
			Name  string `json:"name" validate:"required,min=2,max=150"`
			Email string `json:"email" validate:"required,email,max=150"`
			Phone string `json:"phone" validate:"required,min=8,max=20"`
		}{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "08123456789",
		},
		Items: []CreateItem{{TicketID: ticketMap["ENTRANCE-UMUM"].ID, Quantity: 1}},
	}

	result, err := uc.CreateBooking(ctx, userID, req)
	require.NoError(t, err)

	// Batalkan booking
	err = uc.Cancel(ctx, result.Booking.ID, userID, false)
	require.NoError(t, err)

	// Pastikan status berubah
	booking, err := bookingRepo.FindByID(ctx, result.Booking.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusCancelled, booking.Status)
}

// TestBookingCancelAlreadyPaid menguji bahwa pembatalan booking yang sudah dibayar harus gagal.
func TestBookingCancelAlreadyPaid(t *testing.T) {
	db := setupTestDB(t)
	ticketMap := seedTestTickets(t, db)

	bookingRepo := NewRepository(db)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &testAuditLogger{svc: auditService}
	paymentStarter := &mockPaymentStarter{}

	svc := NewService(bookingRepo, paymentStarter, nil, auditLogger)
	uc := NewUsecase(svc, &testTicketPricer{svc: ticketTypeService, sold: bookingRepo})

	ctx := context.Background()
	userID := "user-123"

	// Buat booking pending dulu
	req := CreateRequest{
		VisitDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		Visitor: struct {
			Name  string `json:"name" validate:"required,min=2,max=150"`
			Email string `json:"email" validate:"required,email,max=150"`
			Phone string `json:"phone" validate:"required,min=8,max=20"`
		}{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "08123456789",
		},
		Items: []CreateItem{{TicketID: ticketMap["ENTRANCE-UMUM"].ID, Quantity: 1}},
	}

	result, err := uc.CreateBooking(ctx, userID, req)
	require.NoError(t, err)

	// Set manual ke PAID
	booking, _ := bookingRepo.FindByID(ctx, result.Booking.ID)
	booking.Status = StatusPaid
	require.NoError(t, bookingRepo.Update(ctx, booking))

	// Coba batalkan, harus gagal
	err = uc.Cancel(ctx, result.Booking.ID, userID, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Booking ini tidak dapat dibatalkan.")
}

// TestBookingSetStatusAdmin menguji override status oleh admin.
func TestBookingSetStatusAdmin(t *testing.T) {
	db := setupTestDB(t)
	ticketMap := seedTestTickets(t, db)

	bookingRepo := NewRepository(db)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &testAuditLogger{svc: auditService}
	paymentStarter := &mockPaymentStarter{}

	svc := NewService(bookingRepo, paymentStarter, nil, auditLogger)
	uc := NewUsecase(svc, &testTicketPricer{svc: ticketTypeService, sold: bookingRepo})

	ctx := context.Background()
	userID := "user-123"

	// Buat booking pending dulu
	req := CreateRequest{
		VisitDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		Visitor: struct {
			Name  string `json:"name" validate:"required,min=2,max=150"`
			Email string `json:"email" validate:"required,email,max=150"`
			Phone string `json:"phone" validate:"required,min=8,max=20"`
		}{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "08123456789",
		},
		Items: []CreateItem{{TicketID: ticketMap["ENTRANCE-UMUM"].ID, Quantity: 1}},
	}

	result, err := uc.CreateBooking(ctx, userID, req)
	require.NoError(t, err)

	// Admin set status ke CONFIRMED
	updated, err := uc.SetStatus(ctx, result.Booking.ID, StatusConfirmed)
	require.NoError(t, err)
	assert.Equal(t, StatusConfirmed, updated.Status)

	// Admin set status ke USED (dari CONFIRMED)
	updated, err = uc.SetStatus(ctx, result.Booking.ID, StatusUsed)
	require.NoError(t, err)
	assert.Equal(t, StatusUsed, updated.Status)

	// Admin coba set EXPIRED dari USED, harus gagal
	_, err = uc.SetStatus(ctx, result.Booking.ID, StatusExpired)
	assert.Error(t, err)
}

// TestBookingSetStatusInvalidTransition menguji transisi status yang tidak valid.
func TestBookingSetStatusInvalidTransition(t *testing.T) {
	db := setupTestDB(t)
	ticketMap := seedTestTickets(t, db)

	bookingRepo := NewRepository(db)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &testAuditLogger{svc: auditService}
	paymentStarter := &mockPaymentStarter{}

	svc := NewService(bookingRepo, paymentStarter, nil, auditLogger)
	uc := NewUsecase(svc, &testTicketPricer{svc: ticketTypeService, sold: bookingRepo})

	ctx := context.Background()
	userID := "user-123"

	// Buat booking pending dulu
	req := CreateRequest{
		VisitDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		Visitor: struct {
			Name  string `json:"name" validate:"required,min=2,max=150"`
			Email string `json:"email" validate:"required,email,max=150"`
			Phone string `json:"phone" validate:"required,min=8,max=20"`
		}{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "08123456789",
		},
		Items: []CreateItem{{TicketID: ticketMap["ENTRANCE-UMUM"].ID, Quantity: 1}},
	}

	result, err := uc.CreateBooking(ctx, userID, req)
	require.NoError(t, err)

	// Coba set PAID langsung dari PENDING tanpa pembayaran, harus gagal
	_, err = uc.SetStatus(ctx, result.Booking.ID, StatusPaid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Status PAID hanya dapat diatur melalui sistem pembayaran.")

	// Coba status tidak valid
	_, err = uc.SetStatus(ctx, result.Booking.ID, "INVALID_STATUS")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Status booking tidak valid.")
}

// TestBookingUnauthorizedAccess menguji bahwa pengguna tidak bisa mengakses booking pengguna lain.
func TestBookingUnauthorizedAccess(t *testing.T) {
	db := setupTestDB(t)
	ticketMap := seedTestTickets(t, db)

	bookingRepo := NewRepository(db)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &testAuditLogger{svc: auditService}
	paymentStarter := &mockPaymentStarter{}

	svc := NewService(bookingRepo, paymentStarter, nil, auditLogger)
	uc := NewUsecase(svc, &testTicketPricer{svc: ticketTypeService, sold: bookingRepo})

	ctx := context.Background()
	userID1 := "user-1"
	userID2 := "user-2"

	// User 1 membuat booking
	req := CreateRequest{
		VisitDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		Visitor: struct {
			Name  string `json:"name" validate:"required,min=2,max=150"`
			Email string `json:"email" validate:"required,email,max=150"`
			Phone string `json:"phone" validate:"required,min=8,max=20"`
		}{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "08123456789",
		},
		Items: []CreateItem{{TicketID: ticketMap["ENTRANCE-UMUM"].ID, Quantity: 1}},
	}

	result, err := uc.CreateBooking(ctx, userID1, req)
	require.NoError(t, err)

	// User 2 coba akses, harus gagal
	_, err = uc.Get(ctx, result.Booking.ID, userID2, false, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Anda tidak memiliki akses.")

	// User 2 coba batalkan, harus gagal
	err = uc.Cancel(ctx, result.Booking.ID, userID2, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Anda tidak memiliki akses.")
}

// Mock payment starter untuk pengujian
type mockPaymentStarter struct {
	payments map[string]*PaymentView
}

func (m *mockPaymentStarter) StartPayment(ctx context.Context, bookingID, visitorName, visitorEmail string, amount int64) (*PaymentView, error) {
	if m.payments == nil {
		m.payments = make(map[string]*PaymentView)
	}
	if existing, ok := m.payments[bookingID]; ok {
		return existing, nil
	}
	p := &PaymentView{
		ID:        "pay-" + bookingID,
		BookingID: bookingID,
		Method:    "QRIS",
		Amount:    amount,
		Status:    "PENDING",
		CreatedAt: time.Now(),
		ExpiredAt: time.Now().Add(30 * time.Minute),
	}
	m.payments[bookingID] = p
	return p, nil
}

func (m *mockPaymentStarter) ExpirePendingByBooking(ctx context.Context, bookingID string) error {
	if p, ok := m.payments[bookingID]; ok {
		p.Status = "EXPIRED"
	}
	return nil
}
