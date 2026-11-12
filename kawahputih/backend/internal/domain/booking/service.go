package booking

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/KawahPutih/backend/internal/domain/audit"
	"github.com/KawahPutih/backend/pkg/apperror"
)

const serviceFeeRate = 2.0

type PaymentStarter interface {
	StartPayment(ctx context.Context, bookingID, visitorName, visitorEmail string, amount int64) (*PaymentView, error)
	ExpirePendingByBooking(ctx context.Context, bookingID string) error
}

type TicketStatusSyncer interface {
	SyncStatusForBooking(ctx context.Context, bookingID, bookingStatus string) error
}

type AuditLogger interface {
	LogSuccess(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string)
	LogFailure(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string)
}

type BookingPaymentUpdater interface {
	MarkPaid(ctx context.Context, bookingID string) error
	MarkExpired(ctx context.Context, bookingID string) error
}

type Service struct {
	repo     Repository
	payments PaymentStarter
	tickets  TicketStatusSyncer
	audit    AuditLogger
}

func NewService(repo Repository, payments PaymentStarter, tickets TicketStatusSyncer, audit AuditLogger) *Service {
	return &Service{repo: repo, payments: payments, tickets: tickets, audit: audit}
}

// TicketPricer menyelesaikan jenis tiket (nama/harga/kapasitas/aktif) berdasarkan ID, dipenuhi service domain tickettype lewat bootstrap.
type TicketPricer interface {
	GetMany(ctx context.Context, ids []string) ([]TicketInfo, error)
	SoldByTicketOn(ctx context.Context, date string) (map[string]int64, error)
}

// TicketInfo meniru field yang dibutuhkan booking dari jenis tiket tanpa mengimpor package tickettype.
type TicketInfo struct {
	ID       string
	Name     string
	Price    int64
	Capacity int
	IsActive bool
}

func (s *Service) Create(ctx context.Context, userID string, pricer TicketPricer, req CreateRequest) (*CreateBookingResult, error) {
	// validasi tanggal kunjungan: bisa diparse dan hari ini atau setelahnya (waktu lokal).
	today := time.Now().Format("2006-01-02")
	if _, err := time.Parse("2006-01-02", req.VisitDate); err != nil || len(req.VisitDate) != 10 {
		return nil, ErrInvalidDate
	}
	if req.VisitDate < today {
		return nil, ErrInvalidDate
	}
	if len(req.Items) == 0 {
		return nil, ErrNoItems
	}
	// gabungkan baris tiket duplikat dan validasi jumlahnya.
	quantities := make(map[string]int)
	ids := make([]string, 0, len(req.Items))
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}
		if _, seen := quantities[item.TicketID]; !seen {
			ids = append(ids, item.TicketID)
		}
		quantities[item.TicketID] += item.Quantity
	}

	// hitung harga tiap baris di server dari katalog.
	types, err := pricer.GetMany(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]TicketInfo, len(types))
	soldByTicket, err := pricer.SoldByTicketOn(ctx, req.VisitDate)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	for _, t := range types {
		if !t.IsActive {
			return nil, unknownTicket(t.ID)
		}
		byID[t.ID] = t
	}

	// kapasitas: jumlah tidak boleh melebihi ketersediaan tersisa.
	for _, id := range ids {
		t := byID[id]
		sold := soldByTicket[id]
		remaining := int64(t.Capacity) - sold - int64(quantities[id])
		if remaining < 0 {
			return nil, ErrUnavailable
		}
	}

	// total (backend adalah satu-satunya sumber kebenaran).
	var subtotal int64
	totalTickets := 0
	items := make([]BookingItem, 0, len(ids))
	for _, id := range ids {
		t := byID[id]
		qty := quantities[id]
		subtotal += t.Price * int64(qty)
		totalTickets += qty
		items = append(items, BookingItem{
			TicketID: t.ID, TicketName: t.Name, Price: t.Price, Quantity: qty,
		})
	}
	serviceFee := int64(float64(subtotal)*serviceFeeRate/100 + 0.5)
	total := subtotal - 0 + serviceFee

	b := &Booking{
		BookingCode:  generateBookingCode(req.VisitDate),
		UserID:       userID,
		VisitorName:  req.Visitor.Name,
		Email:        req.Visitor.Email,
		Phone:        req.Visitor.Phone,
		VisitDate:    req.VisitDate,
		Note:         req.Note,
		Status:       StatusPending,
		Subtotal:     subtotal,
		ServiceFee:   serviceFee,
		Total:        total,
		TotalTickets: totalTickets,
		Items:        items,
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, apperror.Internal(err)
	}

	payment, err := s.payments.StartPayment(ctx, b.ID, b.VisitorName, b.Email, b.Total)
	if err != nil {
		return nil, err
	}

	if s.audit != nil {
		actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
		itemsMeta := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			itemsMeta = append(itemsMeta, map[string]interface{}{
				"ticket_id": item.TicketID, "name": item.TicketName, "price": item.Price, "quantity": item.Quantity,
			})
		}
		s.audit.LogSuccess(ctx, actorID, actorRole, audit.ActionBookingCreate, "booking", b.ID,
			audit.BookingMetadata(b.BookingCode, b.VisitDate, b.Total, itemsMeta), ip, ua)
	}

	return &CreateBookingResult{Booking: toBookingResponse(b, payment), Payment: payment}, nil
}

func (s *Service) authorize(ctx context.Context, id, requesterID string, privileged bool) (*Booking, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if b == nil {
		return nil, ErrNotFound
	}
	if !privileged && b.UserID != requesterID {
		return nil, ErrNotOwner
	}
	return b, nil
}

func (s *Service) Get(ctx context.Context, id, requesterID string, privileged bool, paymentReader PaymentReader) (*BookingResponse, error) {
	b, err := s.authorize(ctx, id, requesterID, privileged)
	if err != nil {
		return nil, err
	}
	var pv *PaymentView
	if paymentReader != nil {
		pv = paymentReader(ctx, b.ID)
	}
	return toBookingResponsePtr(b, pv), nil
}

type PaymentReader func(ctx context.Context, bookingID string) *PaymentView

func (s *Service) ListMine(ctx context.Context, userID string, page, limit int, status string, paymentReader PaymentReader) ([]BookingResponse, int64, error) {
	offset, limit := clampPagination(page, limit)
	bookings, total, err := s.repo.ListMine(ctx, userID, offset, limit, status)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return toBookingResponseList(ctx, bookings, paymentReader), total, nil
}

// ListForUser dipakai panel admin/manager untuk riwayat booking seorang pengunjung.
func (s *Service) ListForUser(ctx context.Context, userID string, page, limit int, status string, paymentReader PaymentReader) ([]BookingResponse, int64, error) {
	offset, limit := clampPagination(page, limit)
	bookings, total, err := s.repo.ListMine(ctx, userID, offset, limit, status)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return toBookingResponseList(ctx, bookings, paymentReader), total, nil
}

func (s *Service) ListAll(ctx context.Context, page, limit int, status, search string, paymentReader PaymentReader) ([]BookingResponse, int64, error) {
	offset, limit := clampPagination(page, limit)
	bookings, total, err := s.repo.ListAll(ctx, offset, limit, status, search)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return toBookingResponseList(ctx, bookings, paymentReader), total, nil
}

func (s *Service) ListByVisitDate(ctx context.Context, date string, page, limit int, paymentReader PaymentReader) ([]BookingResponse, int64, error) {
	offset, limit := clampPagination(page, limit)
	bookings, total, err := s.repo.ListByVisitDate(ctx, date, offset, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return toBookingResponseList(ctx, bookings, paymentReader), total, nil
}

func (s *Service) GetByCode(ctx context.Context, code, requesterID string, privileged bool, paymentReader PaymentReader) (*BookingResponse, error) {
	b, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if b == nil {
		return nil, ErrNotFound
	}
	if !privileged && b.UserID != requesterID {
		return nil, ErrNotOwner
	}
	var pv *PaymentView
	if paymentReader != nil {
		pv = paymentReader(ctx, b.ID)
	}
	return toBookingResponsePtr(b, pv), nil
}

func (s *Service) Cancel(ctx context.Context, id, requesterID string, privileged bool) error {
	b, err := s.authorize(ctx, id, requesterID, privileged)
	if err != nil {
		return err
	}
	if b.Status != StatusPending {
		return ErrCannotCancel
	}
	oldStatus := b.Status
	b.Status = StatusCancelled
	if err := s.repo.Update(ctx, b); err != nil {
		return apperror.Internal(err)
	}
	if err := s.payments.ExpirePendingByBooking(ctx, b.ID); err != nil {
		return err
	}

	if s.audit != nil {
		actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
		s.audit.LogSuccess(ctx, actorID, actorRole, audit.ActionBookingCancel, "booking", b.ID,
			audit.StatusChangeMetadata(oldStatus, b.Status), ip, ua)
	}

	return nil
}

func (s *Service) SetStatus(ctx context.Context, id, status string) (*BookingResponse, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if b == nil {
		return nil, ErrNotFound
	}
	switch status {
	case StatusPending, StatusPaid, StatusConfirmed, StatusUsed, StatusCancelled, StatusExpired:
	default:
		return nil, ErrInvalidStatus
	}
	if b.Status == status {
		return toBookingResponsePtr(b, nil), nil
	}
	if status == StatusUsed && b.Status != StatusPaid && b.Status != StatusConfirmed {
		return nil, apperror.Conflict("Hanya booking yang sudah dibayar yang dapat digunakan.")
	}
	// Admin hanya bisa mengubah ke CONFIRMED dari PENDING (konfirmasi pembayaran offline). Status PAID hanya diisi sistem pembayaran.
	if status == StatusPaid {
		return nil, apperror.Conflict("Status PAID hanya dapat diatur melalui sistem pembayaran.")
	}
	if status == StatusConfirmed && b.Status != StatusPending {
		return nil, apperror.Conflict("Hanya booking yang menunggu pembayaran yang dapat dikonfirmasi.")
	}
	// Cegah transisi tidak valid dari status akhir
	if b.Status == StatusUsed && (status == StatusExpired || status == StatusCancelled) {
		return nil, apperror.Conflict("Booking yang sudah digunakan tidak dapat dikedaluwarsakan atau dibatalkan.")
	}
	if b.Status == StatusExpired && (status == StatusUsed || status == StatusCancelled) {
		return nil, apperror.Conflict("Booking kedaluwarsa tidak dapat digunakan atau dibatalkan.")
	}
	if b.Status == StatusCancelled && (status == StatusUsed || status == StatusExpired) {
		return nil, apperror.Conflict("Booking yang dibatalkan tidak dapat digunakan atau dikedaluwarsakan.")
	}
	oldStatus := b.Status
	b.Status = status
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, apperror.Internal(err)
	}
	if status == StatusCancelled || status == StatusExpired {
		if err := s.payments.ExpirePendingByBooking(ctx, b.ID); err != nil {
			return nil, err
		}
	}
	if s.tickets != nil {
		if err := s.tickets.SyncStatusForBooking(ctx, b.ID, status); err != nil {
			return nil, err
		}
	}

	// Audit log: status booking diubah admin
	if s.audit != nil {
		actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
		s.audit.LogSuccess(ctx, actorID, actorRole, audit.ActionBookingStatus, "booking", b.ID,
			audit.StatusChangeMetadata(oldStatus, b.Status), ip, ua)
	}

	return toBookingResponsePtr(b, nil), nil
}

// MarkPaid dipanggil domain payments saat pembayaran berhasil.
func (s *Service) MarkPaid(ctx context.Context, bookingID string) error {
	b, err := s.repo.FindByID(ctx, bookingID)
	if err != nil {
		return apperror.Internal(err)
	}
	if b == nil {
		return ErrNotFound
	}
	if b.Status == StatusPaid || b.Status == StatusConfirmed || b.Status == StatusUsed {
		return nil // idempotent
	}
	if b.Status != StatusPending {
		return apperror.Conflict("Booking ini sudah tidak dapat dibayar.")
	}
	b.Status = StatusPaid
	if err := s.repo.Update(ctx, b); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) MarkExpired(ctx context.Context, bookingID string) error {
	b, err := s.repo.FindByID(ctx, bookingID)
	if err != nil {
		return apperror.Internal(err)
	}
	if b == nil {
		return ErrNotFound
	}
	if b.Status == StatusExpired || b.Status == StatusPaid || b.Status == StatusUsed {
		return nil
	}
	b.Status = StatusExpired
	if err := s.repo.Update(ctx, b); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) SetCheckedIn(ctx context.Context, bookingID, gate string, at time.Time) error {
	b, err := s.repo.FindByID(ctx, bookingID)
	if err != nil {
		return apperror.Internal(err)
	}
	if b == nil {
		return ErrNotFound
	}
	b.Status = StatusUsed
	b.CheckedInAt = &at
	b.CheckedInGate = gate
	if err := s.repo.Update(ctx, b); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

type Snapshot struct {
	ID          string
	BookingCode string
	VisitorName string
	Email       string
	VisitDate   string
	Status      string
	Total       int64
	Items       []ItemResponse
}

func (s *Service) GetSnapshot(ctx context.Context, bookingID, requesterID string, privileged bool) (*Snapshot, error) {
	b, err := s.authorize(ctx, bookingID, requesterID, privileged)
	if err != nil {
		return nil, err
	}
	snap := &Snapshot{
		ID: b.ID, BookingCode: b.BookingCode, VisitorName: b.VisitorName,
		Email: b.Email, VisitDate: b.VisitDate, Status: b.Status, Total: b.Total,
		Items: make([]ItemResponse, 0, len(b.Items)),
	}
	for _, item := range b.Items {
		snap.Items = append(snap.Items, ItemResponse{TicketID: item.TicketID, TicketName: item.TicketName, Price: item.Price, Quantity: item.Quantity})
	}
	return snap, nil
}

func (s *Service) StalePendingIDs(ctx context.Context, cutoff time.Time, limit int) ([]string, error) {
	bookings, err := s.repo.StalePending(ctx, cutoff, limit)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	ids := make([]string, 0, len(bookings))
	for _, b := range bookings {
		ids = append(ids, b.ID)
	}
	return ids, nil
}

func generateBookingCode(visitDate string) string {
	day := visitDate
	if parsed, err := time.Parse("2006-01-02", visitDate); err == nil {
		day = parsed.Format("20060102")
	}
	buf := make([]byte, 3)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("KP-%s-%s", day, hex.EncodeToString(buf))
}

func clampPagination(page, limit int) (offset, clampedLimit int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return (page - 1) * limit, limit
}

// sortedItemIDs adalah helper untuk output tes/debug yang deterministik.
func sortedItemIDs(m map[string]int) []string {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
