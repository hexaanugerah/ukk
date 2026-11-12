package payments

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
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

type AuditLogger interface {
	LogSuccess(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string)
	LogFailure(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string)
}

type Service struct {
	repo      Repository
	bookings  BookingGateway
	tickets   TicketIssuer
	midtrans  MidtransGateway
	log       *zap.Logger
	audit     AuditLogger
	env       string
	webhookMu sync.Mutex
}

func NewService(repo Repository, bookings BookingGateway, tickets TicketIssuer, midtrans MidtransGateway, log *zap.Logger, audit AuditLogger, environment ...string) *Service {
	env := "production"
	if len(environment) > 0 {
		env = environment[0]
	}
	return &Service{repo: repo, bookings: bookings, tickets: tickets, midtrans: midtrans, log: log, audit: audit, env: env}
}

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
		MidtransOrderID: reference,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		existing, err2 := s.repo.FindLatestByBooking(ctx, bookingID)
		if err2 != nil || existing == nil {
			return nil, apperror.Internal(err)
		}
		return existing, nil
	}

	if s.midtrans != nil {
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

func (s *Service) ExpirePendingByBooking(ctx context.Context, bookingID string) error {
	if _, err := s.repo.ExpirePendingByBooking(ctx, bookingID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) HandleMidtransNotification(ctx context.Context, orderID, transactionStatus, statusCode, fraudStatus, signatureKey, grossAmount string) error {
	if orderID == "" || transactionStatus == "" || statusCode == "" || grossAmount == "" || signatureKey == "" {
		return apperror.Validation("Data notifikasi pembayaran tidak lengkap.")
	}
	if s.midtrans == nil {
		return apperror.Internal(errors.New("midtrans gateway not configured"))
	}
	if !s.midtrans.VerifySignature(orderID, statusCode, grossAmount, signatureKey) {
		return apperror.Unauthorized("Permintaan pembayaran tidak valid.")
	}
	s.webhookMu.Lock()
	defer s.webhookMu.Unlock()

	p, err := s.repo.FindByMidtransOrderID(ctx, orderID)
	if err != nil {
		return apperror.Internal(err)
	}
	if p == nil {
		return apperror.NotFound("Data pembayaran tidak ditemukan.")
	}
	amount, ok := new(big.Rat).SetString(grossAmount)
	if !ok || amount.Cmp(new(big.Rat).SetInt64(p.Amount)) != 0 {
		return apperror.Validation("Nominal notifikasi tidak sesuai dengan pembayaran.")
	}

	snap, err := s.bookings.GetSnapshot(ctx, p.BookingID, "", true)
	if err != nil {
		return err
	}
	if snap == nil {
		return apperror.NotFound("Data booking pembayaran tidak ditemukan.")
	}

	switch transactionStatus {
	case "capture":
		if fraudStatus == "accept" {
			return s.markPaidFromMidtrans(ctx, p, snap)
		}
		if fraudStatus == "challenge" {
			return s.markProcessingFromMidtrans(ctx, p)
		}
		return apperror.Validation("Fraud status notifikasi pembayaran tidak valid.")
	case "settlement":
		return s.markPaidFromMidtrans(ctx, p, snap)
	case "pending":
		return s.markProcessingFromMidtrans(ctx, p)
	case "deny", "cancel", "failure":
		return s.markFailedFromMidtrans(ctx, p, snap)
	case "expire":
		if p.Status == StatusPaid || p.Status == StatusExpired {
			return nil
		}
		_, err := s.expirePayment(ctx, p)
		return err
	default:
		return nil
	}
}

func (s *Service) markPaidFromMidtrans(ctx context.Context, p *Payment, snap *booking.Snapshot) error {
	if p.Status == StatusPaid {
		return nil
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

	if s.audit != nil && snap != nil {
		actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
		s.audit.LogSuccess(ctx, actorID, actorRole, audit.ActionPaymentProcess, "payment", p.ID,
			audit.PaymentMetadata(snap.BookingCode, p.Method, p.Amount, "SUCCESS"), ip, ua)
	}
	return nil
}

func (s *Service) markFailedFromMidtrans(ctx context.Context, p *Payment, snap *booking.Snapshot) error {
	if p.Status == StatusFailed || p.Status == StatusExpired || p.Status == StatusPaid {
		return nil
	}
	p.Status = StatusFailed
	p.PaidAt = nil
	if err := s.repo.Update(ctx, p); err != nil {
		return apperror.Internal(err)
	}
	if s.audit != nil && snap != nil {
		actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
		s.audit.LogFailure(ctx, actorID, actorRole, audit.ActionPaymentProcess, "payment", p.ID,
			audit.PaymentMetadata(snap.BookingCode, p.Method, p.Amount, "FAILED"), ip, ua)
	}
	return nil
}

func (s *Service) markProcessingFromMidtrans(ctx context.Context, p *Payment) error {
	if p.Status == StatusProcessing || p.Status == StatusPaid || p.Status == StatusFailed || p.Status == StatusExpired {
		return nil
	}
	p.Status = StatusProcessing
	return s.repo.Update(ctx, p)
}

func (s *Service) FindByMidtransOrderID(ctx context.Context, orderID string) (*Payment, error) {
	return s.repo.FindByMidtransOrderID(ctx, orderID)
}

func (s *Service) Process(ctx context.Context, bookingID, requesterID string, privileged bool, forced, method, ewalletNumber string) (*booking.BookingResponse, *Payment, string, error) {
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

	if method != "" {
		if !ValidMethod(method) {
			return nil, nil, "", apperror.Validation(fmt.Sprintf("Metode pembayaran %q tidak didukung.", method))
		}
		if ewalletNumber != "" && EwalletNumberFromMethod(method) == "" {
			return nil, nil, "", apperror.Validation("Nomor e-Wallet hanya bisa disematkan untuk metode yang mendukung.")
		}
		pay.Method = method
		pay.EWalletNumber = ewalletNumber
		if err := s.repo.Update(ctx, pay); err != nil {
			return nil, nil, "", apperror.Internal(err)
		}
	}

	if time.Now().After(pay.ExpiredAt) {
		outcome, err := s.expirePayment(ctx, pay)
		if err != nil {
			return nil, nil, "", err
		}
		resp, _ := s.bookings.GetResponse(ctx, bookingID, requesterID, privileged)
		return resp, pay, outcome, nil
	}

	outcome := forced
	if s.env != "local" {
		if forced != "" {
			return nil, nil, "", apperror.Validation("Status pembayaran harus ditentukan oleh Midtrans.")
		}
		if s.midtrans == nil {
			return nil, nil, "", apperror.Internal(errors.New("midtrans gateway not configured"))
		}
		providerStatus, fraudStatus, statusErr := s.midtrans.GetTransactionStatus(ctx, pay.MidtransOrderID)
		if statusErr != nil {
			return nil, nil, "", apperror.Internal(statusErr)
		}
		switch providerStatus {
		case "capture", "settlement":
			if providerStatus == "capture" && fraudStatus != "accept" {
				return nil, nil, "", apperror.Conflict("Pembayaran belum disetujui oleh penyedia.")
			}
			outcome = "SUCCESS"
		case "pending":
			if err := s.markProcessingFromMidtrans(ctx, pay); err != nil {
				return nil, nil, "", err
			}
			resp, err := s.bookings.GetResponse(ctx, bookingID, requesterID, privileged)
			return resp, pay, "PENDING", err
		case "deny", "cancel", "failure":
			outcome = "FAILED"
		case "expire":
			outcome = "EXPIRED"
		default:
			return nil, nil, "", apperror.Conflict("Status pembayaran dari penyedia belum dapat diproses.")
		}
	} else if outcome == "" {
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
				s.log.Error("ticket issuance after payment failed",
					zap.String("booking_id", bookingID), zap.Error(err))
			}
		}
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
	if _, err := s.bookings.GetSnapshot(ctx, pay.BookingID, requesterID, privileged); err != nil {
		return nil, err
	}
	return toPaymentDTOPtr(pay), nil
}

func (s *Service) GetDetail(ctx context.Context, bookingID, requesterID string, privileged bool) (*PaymentDetail, error) {
	pay, err := s.repo.FindLatestByBooking(ctx, bookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if pay == nil {
		return nil, ErrNotFound
	}
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

func (s *Service) MarkDownloaded(ctx context.Context, bookingID, requesterID string, privileged bool) error {
	_, err := s.bookings.GetSnapshot(ctx, bookingID, requesterID, privileged)
	return err
}

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
	default:
		pay.Status = status
		if status != StatusPaid {
			pay.PaidAt = nil
		}
		if err := s.repo.Update(ctx, pay); err != nil {
			return nil, nil, apperror.Internal(err)
		}
	}

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

func (s *Service) FindLatestForBooking(ctx context.Context, bookingID string) (*Payment, error) {
	return s.repo.FindLatestByBooking(ctx, bookingID)
}

func (s *Service) GetByEWallet(ctx context.Context, ewallet string) (*Payment, error) {
	if ewallet == "" {
		return nil, apperror.Validation("Nomor e-Wallet wajib diisi.")
	}
	return s.repo.FindByEWallet(ctx, ewallet)
}

func (s *Service) SetEWalletNumber(ctx context.Context, paymentID, ewalletNumber string) (*Payment, error) {
	pay, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if pay == nil {
		return nil, ErrNotFound
	}
	if ewalletNumber != "" && EwalletNumberFromMethod(pay.Method) != ewalletNumber {
		return nil, apperror.Validation("Nomor e-Wallet tidak sesuai dengan metode pembayaran.")
	}
	pay.EWalletNumber = ewalletNumber
	if err := s.repo.Update(ctx, pay); err != nil {
		return nil, apperror.Internal(err)
	}
	return pay, nil
}

func generateReference() string {
	buf := make([]byte, 6)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("PAY-KP-%s-%s", time.Now().Format("20060102150405"), hex.EncodeToString(buf))
}
