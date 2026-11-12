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
)

// PaymentDTO cocok persis dengan tipe Payment frontend.
type PaymentDTO struct {
	ID        string     `json:"id"`
	BookingID string     `json:"booking_id"`
	Method    string     `json:"method"`
	Amount    int64      `json:"amount"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	PaidAt    *time.Time `json:"paid_at,omitempty"`
	ExpiredAt time.Time  `json:"expired_at"`
}

func toPaymentDTO(p *Payment) PaymentDTO {
	return PaymentDTO{
		ID: p.ID, BookingID: p.BookingID, Method: p.Method, Amount: p.Amount,
		Status: p.Status, CreatedAt: p.CreatedAt, PaidAt: p.PaidAt, ExpiredAt: p.ExpiredAt,
	}
}

func toPaymentDTOPtr(p *Payment) *PaymentDTO {
	dto := toPaymentDTO(p)
	return &dto
}

// PaymentDetail adalah payload gabungan /payments/booking/:id: field Payment di level atas plus objek ticket/booking dalam satu respons.
type PaymentDetail struct {
	PaymentDTO
	Ticket  *ticketing.TicketResponse `json:"ticket,omitempty"`
	Booking *booking.BookingResponse  `json:"booking,omitempty"`
}

// ProcessResult adalah payload /payments/process.
type ProcessResult struct {
	Booking *booking.BookingResponse `json:"booking"`
	Payment PaymentDTO               `json:"payment"`
	Outcome string                   `json:"outcome"` // SUCCESS | FAILED | EXPIRED
}
