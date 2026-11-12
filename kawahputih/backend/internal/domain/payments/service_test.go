package payments

import (
	"context"
	"testing"
	"time"

	"github.com/KawahPutih/backend/internal/domain/audit"
	"github.com/KawahPutih/backend/internal/domain/booking"
	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&Payment{},
		&booking.Booking{},
		&booking.BookingItem{},
		&ticketing.Ticket{},
		&ticketing.ScanRecord{},
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
		VisitDate:    time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
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
	BookingStatusExpired   = booking.StatusExpired
	BookingStatusCancelled = booking.StatusCancelled
)

// noopTicketIssuer adalah penerbit tiket no-op untuk tes.
type noopTicketIssuer struct{}

func (noopTicketIssuer) IssueForBooking(ctx context.Context, snap *booking.Snapshot) (*ticketing.TicketResponse, error) {
	return nil, nil
}
func (noopTicketIssuer) GetByBookingID(ctx context.Context, bookingID string) (*ticketing.TicketResponse, error) {
	return nil, nil
}

// noopMidtransGateway adalah gateway Midtrans no-op untuk tes.
type noopMidtransGateway struct{}

func (noopMidtransGateway) CreateTransaction(ctx context.Context, orderID string, grossAmount int64, customerName, customerEmail, itemName string) (string, error) {
	return "", nil
}
func (noopMidtransGateway) GetTransactionStatus(ctx context.Context, orderID string) (string, string, error) {
	return "", "", nil
}
func (noopMidtransGateway) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	return true
}

func TestPaymentStartIdempotent(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	paymentRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingUsecase{repo: bookingRepo}
	paymentSvc := NewService(paymentRepo, bookingUsecase, noopTicketIssuer{}, noopMidtransGateway{}, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPending)

	// Panggilan pertama
	p1, err := paymentSvc.StartPayment(ctx, testBooking.ID, testBooking.Total)
	require.NoError(t, err)
	require.NotNil(t, p1)
	assert.Equal(t, StatusPending, p1.Status)

	// Panggilan kedua harus mengembalikan pembayaran yang sama
	p2, err := paymentSvc.StartPayment(ctx, testBooking.ID, testBooking.Total)
	require.NoError(t, err)
	assert.Equal(t, p1.ID, p2.ID)
}

func TestPaymentProcessSuccess(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	paymentRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingUsecase{repo: bookingRepo}
	paymentSvc := NewService(paymentRepo, bookingUsecase, noopTicketIssuer{}, noopMidtransGateway{}, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPending)

	// Buat pembayaran pending awal
	p := &Payment{
		BookingID: testBooking.ID, Method: MethodQRIS, Reference: "PAY-TEST-1",
		Amount: testBooking.Total, Status: StatusPending, ExpiredAt: time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, paymentRepo.Create(ctx, p))

	// Proses pembayaran dengan SUCCESS dipaksa dan metode yang dipilih eksplisit
	resp, pay, outcome, err := paymentSvc.Process(ctx, testBooking.ID, "user-1", false, "SUCCESS", MethodGoPay)
	require.NoError(t, err)
	assert.Equal(t, "SUCCESS", outcome)
	assert.Equal(t, StatusPaid, pay.Status)
	assert.NotNil(t, pay.PaidAt)
	assert.Equal(t, BookingStatusPaid, resp.Status)
	assert.Equal(t, MethodGoPay, pay.Method) // chosen method replaces the QRIS default
}

func TestPaymentProcessFailed(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	paymentRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingUsecase{repo: bookingRepo}
	paymentSvc := NewService(paymentRepo, bookingUsecase, noopTicketIssuer{}, noopMidtransGateway{}, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPending)

	p := &Payment{
		BookingID: testBooking.ID, Method: MethodQRIS, Reference: "PAY-TEST-2",
		Amount: testBooking.Total, Status: StatusPending, ExpiredAt: time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, paymentRepo.Create(ctx, p))

	resp, pay, outcome, err := paymentSvc.Process(ctx, testBooking.ID, "user-1", false, "FAILED", "")
	require.NoError(t, err)
	assert.Equal(t, "FAILED", outcome)
	assert.Equal(t, StatusFailed, pay.Status)
	assert.Nil(t, pay.PaidAt)
	assert.Equal(t, BookingStatusPending, resp.Status) // Booking stays pending on failed payment
}

func TestPaymentProcessExpired(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	paymentRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingUsecase{repo: bookingRepo}
	paymentSvc := NewService(paymentRepo, bookingUsecase, noopTicketIssuer{}, noopMidtransGateway{}, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPending)

	// Pembayaran sudah kedaluwarsa
	p := &Payment{
		BookingID: testBooking.ID, Method: MethodQRIS, Reference: "PAY-TEST-3",
		Amount: testBooking.Total, Status: StatusPending, ExpiredAt: time.Now().Add(-1 * time.Minute),
	}
	require.NoError(t, paymentRepo.Create(ctx, p))

	resp, pay, outcome, err := paymentSvc.Process(ctx, testBooking.ID, "user-1", false, "SUCCESS", "")
	require.NoError(t, err)
	assert.Equal(t, "EXPIRED", outcome)
	assert.Equal(t, StatusExpired, pay.Status)
	assert.Equal(t, BookingStatusExpired, resp.Status)
}

func TestPaymentProcessAlreadyPaid(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	paymentRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingUsecase{repo: bookingRepo}
	paymentSvc := NewService(paymentRepo, bookingUsecase, noopTicketIssuer{}, noopMidtransGateway{}, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPaid)

	now := time.Now()
	p := &Payment{
		BookingID: testBooking.ID, Method: MethodQRIS, Reference: "PAY-TEST-4",
		Amount: testBooking.Total, Status: StatusPaid, PaidAt: &now, ExpiredAt: time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, paymentRepo.Create(ctx, p))

	// Coba proses lagi; implementasi sekarang menolak booking non-pending
	_, _, _, err := paymentSvc.Process(ctx, testBooking.ID, "user-1", false, "SUCCESS", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak sedang menunggu pembayaran")
}

func TestPaymentProcessNotPending(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	paymentRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingUsecase{repo: bookingRepo}
	paymentSvc := NewService(paymentRepo, bookingUsecase, noopTicketIssuer{}, noopMidtransGateway{}, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusCancelled)

	p := &Payment{
		BookingID: testBooking.ID, Method: MethodQRIS, Reference: "PAY-TEST-5",
		Amount: testBooking.Total, Status: StatusPending, ExpiredAt: time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, paymentRepo.Create(ctx, p))

	_, _, _, err := paymentSvc.Process(ctx, testBooking.ID, "user-1", false, "SUCCESS", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak sedang menunggu pembayaran")
}

func TestPaymentExpire(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	paymentRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingUsecase{repo: bookingRepo}
	paymentSvc := NewService(paymentRepo, bookingUsecase, noopTicketIssuer{}, noopMidtransGateway{}, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPending)

	p := &Payment{
		BookingID: testBooking.ID, Method: MethodQRIS, Reference: "PAY-TEST-6",
		Amount: testBooking.Total, Status: StatusPending, ExpiredAt: time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, paymentRepo.Create(ctx, p))

	err := paymentSvc.Expire(ctx, testBooking.ID, "user-1", false)
	require.NoError(t, err)

	pay, _ := paymentRepo.FindLatestByBooking(ctx, testBooking.ID)
	assert.Equal(t, StatusExpired, pay.Status)

	b, _ := bookingRepo.FindByID(ctx, testBooking.ID)
	assert.Equal(t, BookingStatusExpired, b.Status)
}

func TestPaymentExpireNonPending(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	paymentRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingUsecase{repo: bookingRepo}
	paymentSvc := NewService(paymentRepo, bookingUsecase, noopTicketIssuer{}, noopMidtransGateway{}, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPaid)

	p := &Payment{
		BookingID: testBooking.ID, Method: MethodQRIS, Reference: "PAY-TEST-7",
		Amount: testBooking.Total, Status: StatusPaid, PaidAt: &[]time.Time{time.Now()}[0], ExpiredAt: time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, paymentRepo.Create(ctx, p))

	err := paymentSvc.Expire(ctx, testBooking.ID, "user-1", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Hanya booking yang menunggu pembayaran yang dapat dikedaluwarsakan.")
}

func TestPaymentSetStatusByAdmin(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	paymentRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingUsecase{repo: bookingRepo}
	paymentSvc := NewService(paymentRepo, bookingUsecase, noopTicketIssuer{}, noopMidtransGateway{}, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPending)

	p := &Payment{
		BookingID: testBooking.ID, Method: MethodQRIS, Reference: "PAY-TEST-8",
		Amount: testBooking.Total, Status: StatusPending, ExpiredAt: time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, paymentRepo.Create(ctx, p))

	// Admin set ke PAID
	pay, resp, err := paymentSvc.SetStatusByAdmin(ctx, p.ID, StatusPaid)
	require.NoError(t, err)
	assert.Equal(t, StatusPaid, pay.Status)
	assert.NotNil(t, pay.PaidAt)
	assert.Equal(t, BookingStatusPaid, resp.Status)

	// Admin set ke REFUNDED
	pay, resp, err = paymentSvc.SetStatusByAdmin(ctx, p.ID, StatusRefunded)
	require.NoError(t, err)
	assert.Equal(t, StatusRefunded, pay.Status)
	assert.Equal(t, BookingStatusCancelled, resp.Status)
}

func TestPaymentSetStatusByAdminInvalidStatus(t *testing.T) {
	db := setupTestDB(t)
	bookingRepo := booking.NewRepository(db)
	paymentRepo := NewRepository(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	bookingUsecase := &mockBookingUsecase{repo: bookingRepo}
	paymentSvc := NewService(paymentRepo, bookingUsecase, noopTicketIssuer{}, noopMidtransGateway{}, nil, auditLogger)

	ctx := context.Background()
	testBooking := createTestBooking(t, db, "user-1", BookingStatusPending)

	p := &Payment{
		BookingID: testBooking.ID, Method: MethodQRIS, Reference: "PAY-TEST-9",
		Amount: testBooking.Total, Status: StatusPending, ExpiredAt: time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, paymentRepo.Create(ctx, p))

	_, _, err := paymentSvc.SetStatusByAdmin(ctx, p.ID, "INVALID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Status pembayaran tidak valid.")
}

// Implementasi mock untuk pengujian

type mockBookingUsecase struct {
	repo booking.Repository
}

func (m *mockBookingUsecase) GetSnapshot(ctx context.Context, bookingID, requesterID string, privileged bool) (*booking.Snapshot, error) {
	b, err := m.repo.FindByID(ctx, bookingID)
	if err != nil || b == nil {
		return nil, apperror.NotFound("booking not found")
	}
	return &booking.Snapshot{
		ID: b.ID, BookingCode: b.BookingCode, VisitorName: b.VisitorName,
		Email: b.Email, VisitDate: b.VisitDate, Status: b.Status, Total: b.Total,
		Items: []booking.ItemResponse{{TicketID: b.Items[0].TicketID, TicketName: b.Items[0].TicketName, Price: b.Items[0].Price, Quantity: b.Items[0].Quantity}},
	}, nil
}

func (m *mockBookingUsecase) MarkPaid(ctx context.Context, bookingID string) error {
	b, err := m.repo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}
	b.Status = BookingStatusPaid
	return m.repo.Update(ctx, b)
}

func (m *mockBookingUsecase) MarkExpired(ctx context.Context, bookingID string) error {
	b, err := m.repo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}
	b.Status = BookingStatusExpired
	return m.repo.Update(ctx, b)
}

func (m *mockBookingUsecase) GetResponse(ctx context.Context, bookingID, requesterID string, privileged bool) (*booking.BookingResponse, error) {
	b, err := m.repo.FindByID(ctx, bookingID)
	if err != nil || b == nil {
		return nil, apperror.NotFound("booking not found")
	}
	createdAt, updatedAt := b.CreatedAt.Format(time.RFC3339), b.UpdatedAt.Format(time.RFC3339)
	return &booking.BookingResponse{
		ID: b.ID, UserID: b.UserID, BookingCode: b.BookingCode,
		VisitorName: b.VisitorName, Email: b.Email, Phone: b.Phone,
		VisitDate: b.VisitDate, Items: toItemResponse(b.Items),
		Subtotal: b.Subtotal, Discount: b.Discount, ServiceFee: b.ServiceFee,
		Tax: b.Tax, Total: b.Total, TotalTickets: b.TotalTickets,
		Status: b.Status, Note: b.Note, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

func (m *mockBookingUsecase) SetStatusByAdmin(ctx context.Context, bookingID, status string) (*booking.BookingResponse, error) {
	b, err := m.repo.FindByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	b.Status = status
	if err := m.repo.Update(ctx, b); err != nil {
		return nil, err
	}
	return m.GetResponse(ctx, bookingID, "", true)
}

func (m *mockBookingUsecase) StalePendingIDs(ctx context.Context, cutoff time.Time, limit int) ([]string, error) {
	bookings, err := m.repo.StalePending(ctx, cutoff, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(bookings))
	for _, b := range bookings {
		ids = append(ids, b.ID)
	}
	return ids, nil
}

func toItemResponse(items []booking.BookingItem) []booking.ItemResponse {
	out := make([]booking.ItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, booking.ItemResponse{
			TicketID: item.TicketID, TicketName: item.TicketName,
			Price: item.Price, Quantity: item.Quantity,
		})
	}
	return out
}

// Menjaga daftar metode dari penyimpangan: setiap konstanta di Methods harus diterima, dan nilai tidak dikenal/kosong tidak boleh. Validator `oneof` di ProcessRequest adalah bagian lain dari kontrak ini.
func TestValidPaymentMethod(t *testing.T) {
	assert.True(t, ValidMethod(MethodQRIS), "the default method created by StartPayment must be valid")
	for _, m := range Methods {
		assert.True(t, ValidMethod(m), "method %q is listed in Methods but rejected by ValidMethod", m)
	}
	assert.False(t, ValidMethod(""))
	assert.False(t, ValidMethod("BITCOIN"))
	assert.False(t, ValidMethod("qris"), "method codes are case-sensitive")
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
