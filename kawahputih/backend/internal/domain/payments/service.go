package payments

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/KawahPutih/backend/internal/domain/audit"
	"github.com/KawahPutih/backend/internal/domain/booking"
	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/pkg/apperror"
	"go.uber.org/zap"
)

type BookingGateway interface {
	GetSnapshot(ctx context.Context, bookingID, requesterID string, privileged bool) (*booking.Snapshot, error)
	MarkPaid(ctx context.Context, bookingID string) error
	MarkExpired(ctx context.Context, bookingID string) error
	GetResponse(ctx context.Context, bookingID, requesterID string, privileged bool) (*booking.BookingResponse, error)
	SetStatusByAdmin(ctx context.Context, bookingID, status string) (*booking.BookingResponse, error)
	StalePendingIDs(ctx context.Context, cutoff time.Time, limit int) ([]string, error)
}

type TicketIssuer interface {
	IssueForBooking(ctx context.Context, snap *booking.Snapshot) (*ticketing.TicketResponse, error)
	GetByBookingID(ctx context.Context, bookingID string) (*ticketing.TicketResponse, error)
}

type MidtransGateway interface {
	CreateTransaction(ctx context.Context, orderID string, grossAmount int64, customerName, customerEmail, itemName string) (string, error)
	GetTransactionStatus(ctx context.Context, orderID string) (status, fraudStatus string, err error)
	VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool
}

// AuditLogger adalah interface minimal yang dibutuhkan payments untuk audit logging.
type AuditLogger interface {
	LogSuccess(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string)
	LogFailure(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string)
}

type Service struct {
	repo     Repository
	bookings BookingGateway
	tickets  TicketIssuer
	midtrans MidtransGateway
	log      *zap.Logger
	audit    AuditLogger
}

func NewService(repo Repository, bookings BookingGateway, tickets TicketIssuer, midtrans MidtransGateway, log *zap.Logger, audit AuditLogger) *Service {
	return &Service{repo: repo, bookings: bookings, tickets: tickets, midtrans: midtrans, log: log, audit: audit}
}

// dipanggil domain booking lewat adapter bootstrap

// StartPayment membuka pembayaran QRIS pending untuk booking yang baru dibuat. Idempoten per booking: bila sudah ada pembayaran yang bisa dibayar, dikembalikan apa adanya.
func (s *Service) StartPayment(ctx context.Context, bookingID string, amount int64) (*Payment, error) {
	existing, err := s.repo.FindLatestByBooking(ctx, bookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if existing != nil && (existing.Status == StatusPending || existing.Status == StatusProcessing) {
		return existing, nil
	}

	reference := generateReference()

	p := &Payment{
		BookingID:       bookingID,
		Method:          MethodQRIS,
		Reference:       reference,
		Amount:          amount,
		Status:          StatusPending,
		ExpiredAt:       time.Now().Add(PaymentWindow),
		MidtransOrderID: reference, // Use reference as Midtrans order ID
	}
	if err := s.repo.Create(ctx, p); err != nil {
		// Unique constraint di booking_id: pembuatan bersamaan menang, jadi baca ulang.
		existing, err2 := s.repo.FindLatestByBooking(ctx, bookingID)
		if err2 != nil || existing == nil {
			return nil, apperror.Internal(err)
		}
		return existing, nil
	}

	// Bila gateway Midtrans tersedia, buat transaksinya
	if s.midtrans != nil {
		// Ambil snapshot booking untuk data pelanggan
		snap, err := s.bookings.GetSnapshot(ctx, bookingID, "", true)
		if err == nil && snap != nil {
			redirectURL, err := s.midtrans.CreateTransaction(ctx, p.MidtransOrderID, amount, snap.VisitorName, snap.Email, "Kawah Putih Booking")
			if err != nil {
				s.log.Error("midtrans create transaction failed", zap.String("booking_id", bookingID), zap.Error(err))
			} else {
				p.MidtransRedirect = redirectURL
				if err := s.repo.Update(ctx, p); err != nil {
					s.log.Error("midtrans redirect url update failed", zap.String("booking_id", bookingID), zap.Error(err))
				}
			}
		}
	}

	return p, nil
}

// ExpirePendingByBooking mengakhiri pembayaran non-final untuk booking (pembatalan booking / override admin).
func (s *Service) ExpirePendingByBooking(ctx context.Context, bookingID string) error {
	if _, err := s.repo.ExpirePendingByBooking(ctx, bookingID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

// HandleMidtransNotification memproses notifikasi webhook Midtrans
func (s *Service) HandleMidtransNotification(ctx context.Context, orderID, transactionStatus, fraudStatus, signatureKey, grossAmount string) error {
	if s.midtrans == nil {
		return apperror.Internal(errors.New("midtrans gateway not configured"))
	}

	// Verifikasi signature
	if !s.midtrans.VerifySignature(orderID, transactionStatus, grossAmount, signatureKey) {
		return apperror.Unauthorized("Permintaan pembayaran tidak valid.")
	}

	p, err := s.repo.FindByMidtransOrderID(ctx, orderID)
	if err != nil {
		return apperror.Internal(err)
	}
	if p == nil {
		return apperror.NotFound("Data pembayaran tidak ditemukan.")
	}

	// Ambil snapshot booking untuk audit
	snap, _ := s.bookings.GetSnapshot(ctx, p.BookingID, "", true)

	switch transactionStatus {
	case "capture", "settlement":
		if fraudStatus == "accept" || fraudStatus == "" {
			return s.markPaidFromMidtrans(ctx, p, snap)
		}
		return s.markFailedFromMidtrans(ctx, p, snap)
	case "pending":
		return s.markProcessingFromMidtrans(ctx, p)
	case "deny", "cancel", "expire":
		return s.markFailedFromMidtrans(ctx, p, snap)
	default:
		return nil // ignore unknown statuses
	}
}

func (s *Service) markPaidFromMidtrans(ctx context.Context, p *Payment, snap *booking.Snapshot) error {
	if p.Status == StatusPaid {
		return nil // idempotent
	}
	now := time.Now()
	p.Status = StatusPaid
	p.PaidAt = &now
	if err := s.repo.Update(ctx, p); err != nil {
		return apperror.Internal(err)
	}
	if err := s.bookings.MarkPaid(ctx, p.BookingID); err != nil {
		return err
	}
	if s.tickets != nil && snap != nil {
		if _, err := s.tickets.IssueForBooking(ctx, snap); err != nil {
			s.log.Error("ticket issuance after midtrans payment failed",
				zap.String("booking_id", p.BookingID), zap.Error(err))
		}
	}

	// Audit log
	if s.audit != nil && snap != nil {
		actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
		s.audit.LogSuccess(ctx, actorID, actorRole, audit.ActionPaymentProcess, "payment", p.ID,
			audit.PaymentMetadata(snap.BookingCode, p.Method, p.Amount, "SUCCESS"), ip, ua)
	}
	return nil
}

func (s *Service) markFailedFromMidtrans(ctx context.Context, p *Payment, snap *booking.Snapshot) error {
	if p.Status == StatusFailed || p.Status == StatusExpired {
		return nil
	}
	p.Status = StatusFailed
	p.PaidAt = nil
	if err := s.repo.Update(ctx, p); err != nil {
		return apperror.Internal(err)
	}

	// Audit log
	if s.audit != nil && snap != nil {
		actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
		s.audit.LogFailure(ctx, actorID, actorRole, audit.ActionPaymentProcess, "payment", p.ID,
			audit.PaymentMetadata(snap.BookingCode, p.Method, p.Amount, "FAILED"), ip, ua)
	}
	return nil
}

func (s *Service) markProcessingFromMidtrans(ctx context.Context, p *Payment) error {
	if p.Status == StatusProcessing || p.Status == StatusPaid {
		return nil
	}
	p.Status = StatusProcessing
	return s.repo.Update(ctx, p)
}

// FindByMidtransOrderID mencari pembayaran berdasarkan order ID Midtrans
func (s *Service) FindByMidtransOrderID(ctx context.Context, orderID string) (*Payment, error) {
	return s.repo.FindByMidtransOrderID(ctx, orderID)
}

// alur untuk pengunjung

// Process menjalankan pembayaran simulasi untuk metode yang dipilih pengunjung. `forced` memungkinkan tes/UI memilih hasil yang pasti; kosong berarti SUCCESS. `method` mencatat metode yang dipakai. Saat berhasil, booking menjadi PAID dan e-tiketnya diterbitkan; kegagalan penerbitan tiket hanya dicatat, tidak membatalkan pembayaran (tiket bisa diterbitkan ulang).
func (s *Service) Process(ctx context.Context, bookingID, requesterID string, privileged bool, forced, method string) (*booking.BookingResponse, *Payment, string, error) {
	snap, err := s.bookings.GetSnapshot(ctx, bookingID, requesterID, privileged)
	if err != nil {
		return nil, nil, "", err
	}
	if snap.Status != booking.StatusPending {
		return nil, nil, "", apperror.Conflict("Booking ini tidak sedang menunggu pembayaran.")
	}
	pay, err := s.repo.FindLatestByBooking(ctx, bookingID)
	if err != nil {
		return nil, nil, "", apperror.Internal(err)
	}
	if pay == nil {
		return nil, nil, "", apperror.NotFound("Data pembayaran tidak ditemukan.")
	}
	if pay.Status == StatusPaid {
		resp, _ := s.bookings.GetResponse(ctx, bookingID, requesterID, privileged)
		return resp, pay, "SUCCESS", nil
	}

	// Catat metode yang dipilih pengunjung sebelum hasil diterapkan agar baris pembayaran mencerminkan yang benar-benar dipakai.
	if method != "" && pay.Method != method {
		pay.Method = method
		if err := s.repo.Update(ctx, pay); err != nil {
			return nil, nil, "", apperror.Internal(err)
		}
	}

	// Jendela yang sudah lewat menang atas hasil yang dipaksa.
	if time.Now().After(pay.ExpiredAt) {
		outcome, err := s.expirePayment(ctx, pay)
		if err != nil {
			return nil, nil, "", err
		}
		resp, _ := s.bookings.GetResponse(ctx, bookingID, requesterID, privileged)
		return resp, pay, outcome, nil
	}

	outcome := forced
	if outcome == "" {
		outcome = "SUCCESS"
	}
	switch outcome {
	case "SUCCESS":
		now := time.Now()
		pay.Status = StatusPaid
		pay.PaidAt = &now
		if err := s.repo.Update(ctx, pay); err != nil {
			return nil, nil, "", apperror.Internal(err)
		}
		if err := s.bookings.MarkPaid(ctx, bookingID); err != nil {
			return nil, nil, "", err
		}
		if s.tickets != nil {
			if _, err := s.tickets.IssueForBooking(ctx, snap); err != nil {
				// Pembayaran berhasil tapi penerbitan tiket gagal; catat dengan jelas karena e-tiket bisa diterbitkan ulang lewat GET.
				s.log.Error("ticket issuance after payment failed",
					zap.String("booking_id", bookingID), zap.Error(err))
			}
		}

		// Audit log: pembayaran berhasil
		if s.audit != nil {
			actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
			s.audit.LogSuccess(ctx, actorID, actorRole, audit.ActionPaymentProcess, "payment", pay.ID,
				audit.PaymentMetadata(snap.BookingCode, pay.Method, pay.Amount, "SUCCESS"), ip, ua)
		}
	case "FAILED":
		pay.Status = StatusFailed
		pay.PaidAt = nil
		if err := s.repo.Update(ctx, pay); err != nil {
			return nil, nil, "", apperror.Internal(err)
		}

		// Audit log: pembayaran gagal
		if s.audit != nil {
			actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
			s.audit.LogFailure(ctx, actorID, actorRole, audit.ActionPaymentProcess, "payment", pay.ID,
				audit.PaymentMetadata(snap.BookingCode, pay.Method, pay.Amount, "FAILED"), ip, ua)
		}
	case "EXPIRED":
		outcome, err = s.expirePayment(ctx, pay)
		if err != nil {
			return nil, nil, "", err
		}

		// Audit log: pembayaran kedaluwarsa
		if s.audit != nil {
			actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
			s.audit.LogFailure(ctx, actorID, actorRole, audit.ActionPaymentProcess, "payment", pay.ID,
				audit.PaymentMetadata(snap.BookingCode, pay.Method, pay.Amount, "EXPIRED"), ip, ua)
		}
	default:
		return nil, nil, "", apperror.Validation("Hasil pembayaran tidak valid.")
	}

	resp, err := s.bookings.GetResponse(ctx, bookingID, requesterID, privileged)
	if err != nil {
		return nil, nil, "", err
	}
	return resp, pay, outcome, nil
}

// Expire mengakhiri pembayaran + booking secara manual (aksi pengunjung atau admin).
func (s *Service) Expire(ctx context.Context, bookingID, requesterID string, privileged bool) error {
	snap, err := s.bookings.GetSnapshot(ctx, bookingID, requesterID, privileged)
	if err != nil {
		return err
	}
	if snap.Status != booking.StatusPending {
		return apperror.Conflict("Hanya booking yang menunggu pembayaran yang dapat dikedaluwarsakan.")
	}
	pay, err := s.repo.FindLatestByBooking(ctx, bookingID)
	if err != nil {
		return apperror.Internal(err)
	}
	if pay == nil {
		return apperror.NotFound("Data pembayaran tidak ditemukan.")
	}
	_, err = s.expirePayment(ctx, pay)

	// Audit log: pembayaran diakhiri manual
	if s.audit != nil && err == nil {
		actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
		s.audit.LogSuccess(ctx, actorID, actorRole, audit.ActionPaymentExpire, "payment", pay.ID,
			audit.PaymentMetadata(snap.BookingCode, pay.Method, pay.Amount, "EXPIRED"), ip, ua)
	}
	return err
}

func (s *Service) expirePayment(ctx context.Context, pay *Payment) (string, error) {
	pay.Status = StatusExpired
	if err := s.repo.Update(ctx, pay); err != nil {
		return "", apperror.Internal(err)
	}
	if err := s.bookings.MarkExpired(ctx, pay.BookingID); err != nil {
		return "", err
	}
	return "EXPIRED", nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID string, privileged bool) (*PaymentDTO, error) {
	pay, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if pay == nil {
		return nil, ErrNotFound
	}
	// Kepemilikan: hanya pemilik pembayaran (lewat booking) atau role berhak yang boleh melihat.
	if _, err := s.bookings.GetSnapshot(ctx, pay.BookingID, requesterID, privileged); err != nil {
		return nil, err
	}
	return toPaymentDTOPtr(pay), nil
}

// GetDetail mengembalikan tampilan gabungan pembayaran untuk /payments/booking/:id: field pembayaran di level atas plus `ticket` dan `booking` dalam satu payload.
func (s *Service) GetDetail(ctx context.Context, bookingID, requesterID string, privileged bool) (*PaymentDetail, error) {
	pay, err := s.repo.FindLatestByBooking(ctx, bookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if pay == nil {
		return nil, ErrNotFound
	}
	// Kepemilikan: hanya pemilik booking atau role berhak yang boleh melihat.
	if _, err := s.bookings.GetSnapshot(ctx, bookingID, requesterID, privileged); err != nil {
		return nil, err
	}
	detail := &PaymentDetail{toPaymentDTO(pay), nil, nil}
	if s.tickets != nil {
		if t, _ := s.tickets.GetByBookingID(ctx, bookingID); t != nil {
			detail.Ticket = t
		}
	}
	if resp, err := s.bookings.GetResponse(ctx, bookingID, requesterID, privileged); err == nil {
		detail.Booking = resp
	}
	return detail, nil
}

// MarkDownloaded mencatat bahwa pengunjung mengunduh e-tiket. PDF-nya dibuat di sisi klien; ini hanya hook audit.
func (s *Service) MarkDownloaded(ctx context.Context, bookingID, requesterID string, privileged bool) error {
	_, err := s.bookings.GetSnapshot(ctx, bookingID, requesterID, privileged)
	return err
}

// admin

func (s *Service) ListAll(ctx context.Context, page, limit int, status string) ([]PaymentDTO, int64, error) {
	offset := (page - 1) * limit
	if page < 1 {
		offset = 0
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	payments, total, err := s.repo.ListAll(ctx, offset, limit, status)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	out := make([]PaymentDTO, 0, len(payments))
	for i := range payments {
		out = append(out, toPaymentDTO(&payments[i]))
	}
	return out, total, nil
}

// SetStatusByAdmin menerapkan override status pembayaran oleh admin dengan efek samping booking/tiket yang konsisten.
func (s *Service) SetStatusByAdmin(ctx context.Context, paymentID, status string) (*Payment, *booking.BookingResponse, error) {
	pay, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, nil, apperror.Internal(err)
	}
	if pay == nil {
		return nil, nil, ErrNotFound
	}
	switch status {
	case StatusPaid, StatusFailed, StatusExpired, StatusRefunded, StatusPending, StatusProcessing:
	default:
		return nil, nil, ErrInvalidStatus
	}

	var snap *booking.Snapshot
	if s.audit != nil {
		snap, _ = s.bookings.GetSnapshot(ctx, pay.BookingID, "", true)
	}

	switch status {
	case StatusPaid:
		if pay.Status != StatusPaid {
			now := time.Now()
			pay.Status = StatusPaid
			pay.PaidAt = &now
			if err := s.repo.Update(ctx, pay); err != nil {
				return nil, nil, apperror.Internal(err)
			}
			if err := s.bookings.MarkPaid(ctx, pay.BookingID); err != nil {
				return nil, nil, err
			}
			if s.tickets != nil {
				if snap, err := s.bookings.GetSnapshot(ctx, pay.BookingID, "", true); err == nil {
					if _, err := s.tickets.IssueForBooking(ctx, snap); err != nil {
						s.log.Error("ticket issuance after admin payment override failed",
							zap.String("booking_id", pay.BookingID), zap.Error(err))
					}
				}
			}
		}
	case StatusRefunded:
		pay.Status = StatusRefunded
		if err := s.repo.Update(ctx, pay); err != nil {
			return nil, nil, apperror.Internal(err)
		}
		resp, err := s.bookings.SetStatusByAdmin(ctx, pay.BookingID, booking.StatusCancelled)
		if err != nil {
			return nil, nil, err
		}
		return pay, resp, nil
	case StatusExpired:
		if _, err := s.expirePayment(ctx, pay); err != nil {
			return nil, nil, err
		}
	default: // FAILED / PENDING / PROCESSING
		pay.Status = status
		if status != StatusPaid {
			pay.PaidAt = nil
		}
		if err := s.repo.Update(ctx, pay); err != nil {
			return nil, nil, apperror.Internal(err)
		}
	}

	// Audit log: status pembayaran diubah admin
	if s.audit != nil && snap != nil {
		actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
		s.audit.LogSuccess(ctx, actorID, actorRole, audit.ActionPaymentStatus, "payment", pay.ID,
			audit.PaymentMetadata(snap.BookingCode, pay.Method, pay.Amount, status), ip, ua)
	}

	resp, err := s.bookings.GetResponse(ctx, pay.BookingID, "", true)
	if err != nil {
		return nil, nil, err
	}
	return pay, resp, nil
}

// ExpireStalePending adalah hook scheduler: mengakhiri pembayaran (dan booking-nya) yang jendelanya lewat. Mengembalikan jumlah yang disapu.
func (s *Service) ExpireStalePending(ctx context.Context, cutoff time.Time, limit int) (int, error) {
	ids, err := s.bookings.StalePendingIDs(ctx, cutoff, limit)
	if err != nil {
		return 0, err
	}
	swept := 0
	for _, id := range ids {
		pay, err := s.repo.FindLatestByBooking(ctx, id)
		if err != nil {
			return swept, err
		}
		if pay != nil && (pay.Status == StatusPending || pay.Status == StatusProcessing || pay.Status == StatusFailed) {
			pay.Status = StatusExpired
			if err := s.repo.Update(ctx, pay); err != nil {
				return swept, err
			}
		}
		if err := s.bookings.MarkExpired(ctx, id); err != nil {
			return swept, err
		}
		swept++
	}
	return swept, nil
}

// FindLatestForBooking membuka pencarian pembayaran mentah untuk adapter pengayaan daftar.
func (s *Service) FindLatestForBooking(ctx context.Context, bookingID string) (*Payment, error) {
	return s.repo.FindLatestByBooking(ctx, bookingID)
}

// generateReference menghasilkan referensi pembayaran (order id ala Midtrans).
func generateReference() string {
	buf := make([]byte, 6)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("PAY-KP-%s-%s", time.Now().Format("20060102150405"), hex.EncodeToString(buf))
}
