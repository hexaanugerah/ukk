package ticketing

import (
	"context"
	"testing"
	"time"

	"github.com/KawahPutih/backend/internal/domain/audit"
	"github.com/KawahPutih/backend/internal/domain/booking"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&Ticket{},
		&ScanRecord{},
		&booking.Booking{},
		&booking.BookingItem{},
		&audit.AuditLog{},
	)
	require.NoError(t, err)

	return db
}

func createTestBooking(t *testing.T, db *gorm.DB, userID, status string) *booking.Booking {
	b := &booking.Booking{
		BookingCode:  "KP-TEST-" + time.Now().Format("20060102-150405"),
		UserID:       userID,
		VisitorName:  "John Doe",
		Email:        "john@example.com",
		Phone:        "08123456789",
		VisitDate:    time.Now().Format("2006-01-02"),
		Status:       status,
		Subtotal:     15000,
		ServiceFee:   300,
		Total:        15300,
		TotalTickets: 1,
		Items: []booking.BookingItem{
			{TicketID: "ENTRANCE-UMUM", TicketName: "Tiket Masuk Umum", Price: 15000, Quantity: 1},
		},
	}
	require.NoError(t, db.Create(b).Error)
	return b
}

// bookingStatus memberi akses konstanta status booking tanpa menutupi nama lain
var (
	BookingStatusPending   = booking.StatusPending
	BookingStatusPaid      = booking.StatusPaid
	BookingStatusUsed      = booking.StatusUsed
	BookingStatusCancelled = booking.StatusCancelled
	BookingStatusExpired   = booking.StatusExpired
)

func createTestTicket(t *testing.T, db *gorm.DB, bookingID, status string) *Ticket {
	qr := "KPT-TEST-" + time.Now().Format("20060102150405")
	tk := &Ticket{
		BookingID:   bookingID,
		VisitorName: "John Doe",
		Email:       "john@example.com",
		VisitDate:   time.Now().Format("2006-01-02"),
		TicketName:  "Tiket Masuk Umum x1",
		Quantity:    1,
		Total:       15300,
		Status:      status,
		GeneratedAt: time.Now(),
		QRData:      qr,
	}
	require.NoError(t, db.Create(tk).Error)
	return tk
}

func TestTicketIssueIdempotent(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	ticketRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingCheckinUsecase{repo: bookingRepo}
	ticketSvc := NewService(ticketRepo, bookingUsecase, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPaid)

	// Penerbitan pertama
	snap := &booking.Snapshot{
		ID: testBooking.ID, BookingCode: testBooking.BookingCode, VisitorName: testBooking.VisitorName,
		Email: testBooking.Email, VisitDate: testBooking.VisitDate, Status: testBooking.Status, Total: testBooking.Total,
		Items: []booking.ItemResponse{{TicketID: "ENTRANCE-UMUM", TicketName: "Tiket Masuk Umum", Price: 15000, Quantity: 1}},
	}
	t1, err := ticketSvc.IssueForBooking(ctx, snap)
	require.NoError(t, err)
	require.NotNil(t, t1)
	assert.Equal(t, TicketValid, t1.Status)

	// Penerbitan kedua harus mengembalikan tiket yang sama
	t2, err := ticketSvc.IssueForBooking(ctx, snap)
	require.NoError(t, err)
	assert.Equal(t, t1.ID, t2.ID)
}

func TestTicketValidateSuccess(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	ticketRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingCheckinUsecase{repo: bookingRepo}
	ticketSvc := NewService(ticketRepo, bookingUsecase, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPaid)

	// Buat tiket valid
	ticket := createTestTicket(t, db, testBooking.ID, TicketValid)

	result, err := ticketSvc.Validate(ctx, ticket.QRData, "Gerbang 1", "petugas-1", "Petugas 1")
	require.NoError(t, err)
	assert.Equal(t, ScanValid, result.Status)
	assert.Equal(t, "Tiket valid, pengunjung dipersilakan masuk.", result.Message)
	assert.Equal(t, TicketUsed, result.Ticket.Status)
	assert.NotNil(t, result.Ticket.CheckedInAt)

	// Pastikan booking ditandai USED
	b, _ := bookingRepo.FindByID(ctx, testBooking.ID)
	assert.Equal(t, BookingStatusUsed, b.Status)
}

// E-tiket lama mengodekan payload JSON berisi id tiket/booking di dalam QR.
// Scanner harus tetap bisa menyelesaikannya menjadi tiket yang sama.
func TestTicketValidateJSONQRPayload(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	ticketRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingCheckinUsecase{repo: bookingRepo}
	ticketSvc := NewService(ticketRepo, bookingUsecase, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPaid)
	ticket := createTestTicket(t, db, testBooking.ID, TicketValid)

	payload := `{"t":"` + ticket.ID + `","b":"` + testBooking.ID + `"}`
	result, err := ticketSvc.Validate(ctx, payload, "Gerbang 1", "petugas-1", "Petugas 1")
	require.NoError(t, err)
	assert.Equal(t, ScanValid, result.Status)
	assert.Equal(t, ticket.ID, result.Ticket.ID)
	assert.Equal(t, TicketUsed, result.Ticket.Status)

	// Riwayat aktivitas menyimpan hasil validasi.
	var scans []ScanRecord
	require.NoError(t, db.Find(&scans).Error)
	require.Len(t, scans, 1)
	assert.Equal(t, ScanValid, scans[0].Status)
	assert.Equal(t, "petugas-1", scans[0].PetugasID)
}

func TestTicketValidateAlreadyUsed(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	ticketRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingCheckinUsecase{repo: bookingRepo}
	ticketSvc := NewService(ticketRepo, bookingUsecase, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusUsed)

	// Buat tiket yang sudah terpakai
	ticket := createTestTicket(t, db, testBooking.ID, TicketUsed)

	result, err := ticketSvc.Validate(ctx, ticket.QRData, "Gerbang 1", "petugas-1", "Petugas 1")
	require.NoError(t, err)
	assert.Equal(t, ScanUsed, result.Status)
	assert.Equal(t, "Tiket sudah digunakan.", result.Message)
}

func TestTicketValidateCancelled(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	ticketRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingCheckinUsecase{repo: bookingRepo}
	ticketSvc := NewService(ticketRepo, bookingUsecase, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusCancelled)

	ticket := createTestTicket(t, db, testBooking.ID, TicketCancelled)

	result, err := ticketSvc.Validate(ctx, ticket.QRData, "Gerbang 1", "petugas-1", "Petugas 1")
	require.NoError(t, err)
	assert.Equal(t, ScanCancelled, result.Status)
	assert.Equal(t, "Tiket dibatalkan.", result.Message)
}

func TestTicketValidateExpired(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	ticketRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingCheckinUsecase{repo: bookingRepo}
	ticketSvc := NewService(ticketRepo, bookingUsecase, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusExpired)

	ticket := createTestTicket(t, db, testBooking.ID, TicketExpired)

	result, err := ticketSvc.Validate(ctx, ticket.QRData, "Gerbang 1", "petugas-1", "Petugas 1")
	require.NoError(t, err)
	assert.Equal(t, ScanExpired, result.Status)
	assert.Equal(t, "Tiket sudah kedaluwarsa.", result.Message)
}

func TestTicketValidateWrongDate(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	ticketRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingCheckinUsecase{repo: bookingRepo}
	ticketSvc := NewService(ticketRepo, bookingUsecase, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPaid)

	// Buat tiket untuk besok
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	ticket := &Ticket{
		BookingID:   testBooking.ID,
		VisitorName: "John Doe",
		Email:       "john@example.com",
		VisitDate:   tomorrow,
		TicketName:  "Tiket Masuk Umum x1",
		Quantity:    1,
		Total:       15300,
		Status:      TicketValid,
		GeneratedAt: time.Now(),
		QRData:      "KPT-TEST-" + time.Now().Format("20060102150405"),
	}
	require.NoError(t, db.Create(ticket).Error)

	result, err := ticketSvc.Validate(ctx, ticket.QRData, "Gerbang 1", "petugas-1", "Petugas 1")
	require.NoError(t, err)
	assert.Equal(t, ScanInvalid, result.Status)
	assert.Contains(t, result.Message, "bukan hari ini")
}

func TestTicketValidateNotFound(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	ticketRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingCheckinUsecase{repo: bookingRepo}
	ticketSvc := NewService(ticketRepo, bookingUsecase, nil, auditLogger)

	ctx := context.Background()

	result, err := ticketSvc.Validate(ctx, "NONEXISTENT", "Gerbang 1", "petugas-1", "Petugas 1")
	require.NoError(t, err)
	assert.Equal(t, ScanInvalid, result.Status)
	assert.Equal(t, "Data tiket tidak ditemukan.", result.Message)
}

func TestTicketValidateByID(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	ticketRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingCheckinUsecase{repo: bookingRepo}
	ticketSvc := NewService(ticketRepo, bookingUsecase, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPaid)
	ticket := createTestTicket(t, db, testBooking.ID, TicketValid)

	// Validasi memakai ID tiket bukan QR
	result, err := ticketSvc.Validate(ctx, ticket.ID, "Gerbang 1", "petugas-1", "Petugas 1")
	require.NoError(t, err)
	assert.Equal(t, ScanValid, result.Status)
}

func TestTicketValidateConcurrentNoDoubleUse(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	ticketRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingCheckinUsecase{repo: bookingRepo}
	ticketSvc := NewService(ticketRepo, bookingUsecase, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPaid)
	ticket := createTestTicket(t, db, testBooking.ID, TicketValid)

	// Validasi pertama berhasil
	result1, err := ticketSvc.Validate(ctx, ticket.QRData, "Gerbang 1", "petugas-1", "Petugas 1")
	require.NoError(t, err)
	assert.Equal(t, ScanValid, result1.Status)

	// Validasi kedua harus mengembalikan USED
	result2, err := ticketSvc.Validate(ctx, ticket.QRData, "Gerbang 1", "petugas-1", "Petugas 1")
	require.NoError(t, err)
	assert.Equal(t, ScanUsed, result2.Status)
}

func TestTicketSyncStatus(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	ticketRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingCheckinUsecase{repo: bookingRepo}
	ticketSvc := NewService(ticketRepo, bookingUsecase, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPaid)
	createTestTicket(t, db, testBooking.ID, TicketValid)

	// Sinkronkan ke CANCELLED
	err := ticketSvc.SyncStatusForBooking(ctx, testBooking.ID, BookingStatusCancelled)
	require.NoError(t, err)

	ticket, _ := ticketRepo.FindByBookingID(ctx, testBooking.ID)
	assert.Equal(t, TicketCancelled, ticket.Status)
}

// Implementasi mock

type mockBookingCheckinUsecase struct {
	repo booking.Repository
}

func (m *mockBookingCheckinUsecase) IsOwner(ctx context.Context, bookingID, userID string) (bool, error) {
	b, err := m.repo.FindByID(ctx, bookingID)
	if err != nil {
		return false, err
	}
	return b != nil && b.UserID == userID, nil
}

func (m *mockBookingCheckinUsecase) SetCheckedIn(ctx context.Context, bookingID, gate string, at time.Time) error {
	b, err := m.repo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}
	b.Status = BookingStatusUsed
	b.CheckedInAt = &at
	b.CheckedInGate = gate
	return m.repo.Update(ctx, b)
}

type auditLoggerAdapter struct {
	svc *audit.Service
}

func (a *auditLoggerAdapter) LogSuccess(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string) {
	a.svc.LogSuccess(ctx, actorID, actorRole, action, targetType, targetID, metadata, ip, ua)
}

func (a *auditLoggerAdapter) LogFailure(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string) {
	a.svc.LogFailure(ctx, actorID, actorRole, action, targetType, targetID, metadata, ip, ua)
}
