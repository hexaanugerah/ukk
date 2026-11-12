package bootstrap

import (
	"context"
	"errors"
	"time"

	"github.com/KawahPutih/backend/internal/domain/audit"
	"github.com/KawahPutih/backend/internal/domain/booking"
	"github.com/KawahPutih/backend/internal/domain/payments"
	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/internal/domain/tickettype"
)

// booking ← tickettype: harga & kapasitas dihitung server

// ticketPricerAdapter memenuhi booking.TicketPricer. Data tiket dari service tickettype; jumlah terjual diambil langsung dari repository booking.
type ticketPricerAdapter struct {
	svc  *tickettype.Service
	sold booking.Repository
}

func (a *ticketPricerAdapter) GetMany(ctx context.Context, ids []string) ([]booking.TicketInfo, error) {
	types, err := a.svc.GetMany(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]booking.TicketInfo, 0, len(types))
	for _, t := range types {
		out = append(out, booking.TicketInfo{
			ID: t.ID, Name: t.Name, Price: t.Price, Capacity: t.Capacity, IsActive: t.IsActive,
		})
	}
	return out, nil
}

func (a *ticketPricerAdapter) SoldByTicketOn(ctx context.Context, date string) (map[string]int64, error) {
	return a.sold.SoldByTicketOn(ctx, date)
}

// booking → payments: hook siklus pembayaran

type paymentStarterAdapter struct {
	svc *payments.Service
}

func (a *paymentStarterAdapter) StartPayment(ctx context.Context, bookingID, visitorName, visitorEmail string, amount int64) (*booking.PaymentView, error) {
	p, err := a.svc.StartPayment(ctx, bookingID, amount)
	if err != nil {
		return nil, err
	}
	return toPaymentView(p), nil
}

func (a *paymentStarterAdapter) ExpirePendingByBooking(ctx context.Context, bookingID string) error {
	return a.svc.ExpirePendingByBooking(ctx, bookingID)
}

func toPaymentView(p *payments.Payment) *booking.PaymentView {
	if p == nil {
		return nil
	}
	return &booking.PaymentView{
		ID: p.ID, BookingID: p.BookingID, Method: p.Method, Amount: p.Amount,
		Status: p.Status, CreatedAt: p.CreatedAt, ExpiredAt: p.ExpiredAt,
	}
}

// bookingPaymentReader menambahkan snapshot pembayaran terbaru ke respons booking.
func bookingPaymentReader(svc *payments.Service) booking.PaymentReader {
	return func(ctx context.Context, bookingID string) *booking.PaymentView {
		p, err := svc.FindLatestForBooking(ctx, bookingID)
		if err != nil {
			return nil
		}
		return toPaymentView(p)
	}
}

// booking → ticketing: jaga status e-tiket tetap sinkron dengan override admin

type ticketStatusSyncerAdapter struct {
	svc *ticketing.Service
}

func (a *ticketStatusSyncerAdapter) SyncStatusForBooking(ctx context.Context, bookingID, bookingStatus string) error {
	return a.svc.SyncStatusForBooking(ctx, bookingID, bookingStatus)
}

// ticketing → booking: scan gerbang menandai booking USED

type bookingCheckinAdapter struct {
	uc *booking.Usecase
}

func (a *bookingCheckinAdapter) SetCheckedIn(ctx context.Context, bookingID, gate string, at time.Time) error {
	return a.uc.SetCheckedIn(ctx, bookingID, gate, at)
}

// IsOwner dipakai ticketing untuk membatasi akses e-tiket milik pengunjung lain.
func (a *bookingCheckinAdapter) IsOwner(ctx context.Context, bookingID, userID string) (bool, error) {
	if userID == "" {
		return false, nil
	}
	_, err := a.uc.GetSnapshot(ctx, bookingID, userID, false)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, booking.ErrNotFound) || errors.Is(err, booking.ErrNotOwner) {
		return false, nil
	}
	return false, err
}

// payments → booking: perubahan status + pembacaan

type bookingGatewayAdapter struct {
	uc *booking.Usecase
}

func (a *bookingGatewayAdapter) GetSnapshot(ctx context.Context, bookingID, requesterID string, privileged bool) (*booking.Snapshot, error) {
	return a.uc.GetSnapshot(ctx, bookingID, requesterID, privileged)
}

func (a *bookingGatewayAdapter) MarkPaid(ctx context.Context, bookingID string) error {
	return a.uc.MarkPaid(ctx, bookingID)
}

func (a *bookingGatewayAdapter) MarkExpired(ctx context.Context, bookingID string) error {
	return a.uc.MarkExpired(ctx, bookingID)
}

func (a *bookingGatewayAdapter) GetResponse(ctx context.Context, bookingID, requesterID string, privileged bool) (*booking.BookingResponse, error) {
	return a.uc.GetResponse(ctx, bookingID, requesterID, privileged)
}

func (a *bookingGatewayAdapter) SetStatusByAdmin(ctx context.Context, bookingID, status string) (*booking.BookingResponse, error) {
	return a.uc.SetStatus(ctx, bookingID, status)
}

func (a *bookingGatewayAdapter) StalePendingIDs(ctx context.Context, cutoff time.Time, limit int) ([]string, error) {
	return a.uc.StalePendingIDs(ctx, cutoff, limit)
}

// payments → ticketing: terbitkan e-tiket saat pembayaran berhasil

type ticketIssuerAdapter struct {
	svc *ticketing.Service
}

func (a *ticketIssuerAdapter) IssueForBooking(ctx context.Context, snap *booking.Snapshot) (*ticketing.TicketResponse, error) {
	return a.svc.IssueForBooking(ctx, snap)
}

func (a *ticketIssuerAdapter) GetByBookingID(ctx context.Context, bookingID string) (*ticketing.TicketResponse, error) {
	return a.svc.GetByBookingID(ctx, bookingID)
}

// tickettype ← booking: ketersediaan di katalog publik

type bookingSalesAdapter struct {
	lag booking.Repository
}

func (a *bookingSalesAdapter) SoldByTicketOn(ctx context.Context, date string) map[string]int64 {
	m, err := a.lag.SoldByTicketOn(ctx, date)
	if err != nil {
		return nil
	}
	return m
}

// Adapter audit: menyambungkan audit logging ke service

type auditLoggerAdapter struct {
	svc *audit.Service
}

func (a *auditLoggerAdapter) LogSuccess(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string) {
	a.svc.LogSuccess(ctx, actorID, actorRole, action, targetType, targetID, metadata, ip, ua)
}

func (a *auditLoggerAdapter) LogFailure(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string) {
	a.svc.LogFailure(ctx, actorID, actorRole, action, targetType, targetID, metadata, ip, ua)
}

// Pengecekan saat kompilasi bahwa auditLoggerAdapter memenuhi interface yang dibutuhkan.
var _ booking.AuditLogger = (*auditLoggerAdapter)(nil)
var _ payments.AuditLogger = (*auditLoggerAdapter)(nil)
var _ ticketing.AuditLogger = (*auditLoggerAdapter)(nil)
