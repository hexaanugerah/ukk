package payments

import (
	"time"

	"github.com/KawahPutih/backend/internal/domain/booking"
	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/pkg/apperror"
)

var (
	ErrNotFound      = apperror.NotFound("Data pembayaran tidak ditemukan.").WithCode("PAY_001")
	ErrInvalidStatus = apperror.Validation("Status pembayaran tidak valid.").WithCode("PAY_002")
	ErrInvalidMethod = apperror.Validation("Metode pembayaran tidak didukung.").WithCode("PAY_003")
)

type PaymentDTO struct {
	ID            string     `json:"id"`
	BookingID     string     `json:"booking_id"`
	Method        string     `json:"method"`
	EWalletNumber string     `json:"ewallet_number,omitempty"`
	Amount        int64      `json:"amount"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	ExpiredAt     time.Time  `json:"expired_at"`
}

func toPaymentDTO(p *Payment) PaymentDTO {
	return PaymentDTO{
		ID:            p.ID,
		BookingID:     p.BookingID,
		Method:        p.Method,
		EWalletNumber: p.EWalletNumber,
		Amount:        p.Amount,
		Status:        p.Status,
		CreatedAt:     p.CreatedAt,
		PaidAt:        p.PaidAt,
		ExpiredAt:     p.ExpiredAt,
	}
}

func toPaymentDTOPtr(p *Payment) *PaymentDTO {
	dto := toPaymentDTO(p)
	return &dto
}

type PaymentDetail struct {
	PaymentDTO
	Ticket  *ticketing.TicketResponse `json:"ticket,omitempty"`
	Booking *booking.BookingResponse  `json:"booking,omitempty"`
}

type ProcessResult struct {
	Booking *booking.BookingResponse `json:"booking"`
	Payment PaymentDTO               `json:"payment"`
	Outcome string                   `json:"outcome"`
}

type ProcessRequest struct {
	BookingID     string `json:"booking_id" validate:"required"`
	Forced        string `json:"forced" validate:"omitempty,oneof=SUCCESS FAILED EXPIRED"`
	Method        string `json:"method" validate:"omitempty"`
	EWalletNumber string `json:"ewallet_number,omitempty" validate:"omitempty,min=1,max=50"`
}

type SetStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=PAID FAILED EXPIRED REFUNDED PENDING PROCESSING"`
}

type SetEWalletRequest struct {
	EWalletNumber string `json:"ewallet_number" validate:"required,min=1,max=50"`
}

type AdminListResponse struct {
	Payments []PaymentDTO `json:"payments"`
	Total    int64        `json:"total"`
}
