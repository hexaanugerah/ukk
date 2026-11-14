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
	ID               string     `json:"id"`
	BookingID        string     `json:"booking_id"`
	Provider         string     `json:"provider"`
	Method           string     `json:"method"`
	Reference        string     `json:"reference,omitempty"`
	EWalletNumber    string     `json:"ewallet_number,omitempty"`
	Amount           int64      `json:"amount"`
	Status           string     `json:"status"`
	SettlementStatus string     `json:"settlement_status"`
	ProofURL         string     `json:"proof_url,omitempty"`
	ProofUploadedAt  *time.Time `json:"proof_uploaded_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	ExpiredAt        time.Time  `json:"expired_at"`
	// QrisString hanya diisi untuk pembayaran yang masih menunggu pembayaran
	// (PENDING/PROCESSING) dan metode QRIS.
	QrisString string `json:"qris_string,omitempty"`
	// MidtransRedirect hanya diisi untuk pembayaran yang masih menunggu
	// pembayaran (PENDING/PROCESSING).
	MidtransRedirect string `json:"midtrans_redirect_url,omitempty"`
}

func toPaymentDTO(p *Payment) PaymentDTO {
	provider := p.Provider
	if provider == "" {
		provider = ProviderMidtrans
	}
	settlement := p.SettlementStatus
	if settlement == "" {
		settlement = SettlementUnsettled
	}
	dto := PaymentDTO{
		ID:               p.ID,
		BookingID:        p.BookingID,
		Provider:         provider,
		Method:           p.Method,
		Reference:        p.Reference,
		EWalletNumber:    p.EWalletNumber,
		Amount:           p.Amount,
		Status:           p.Status,
		SettlementStatus: settlement,
		ProofURL:         p.ProofURL,
		ProofUploadedAt:  p.ProofUploadedAt,
		CreatedAt:        p.CreatedAt,
		PaidAt:           p.PaidAt,
		ExpiredAt:        p.ExpiredAt,
	}
	if p.Status == StatusPending || p.Status == StatusProcessing {
		dto.QrisString = p.QrisString
		dto.MidtransRedirect = p.MidtransRedirect
	}
	return dto
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

type UpdateSettingsRequest struct {
	Enabled                 *bool  `json:"enabled,omitempty"`
	MerchantName            string `json:"merchant_name,omitempty" validate:"omitempty,max=150"`
	SettlementBank          string `json:"settlement_bank,omitempty" validate:"omitempty,max=50"`
	SettlementAccountMasked string `json:"settlement_account_masked,omitempty" validate:"omitempty,max=50"`
}

type UpdateMethodRequest struct {
	Enabled      *bool `json:"enabled"`
	DisplayOrder *int  `json:"display_order,omitempty"`
}

type AdminListResponse struct {
	Payments []PaymentDTO `json:"payments"`
	Total    int64        `json:"total"`
}

// MethodTotal adalah agregat pendapatan per metode pembayaran.
type MethodTotal struct {
	Method string `json:"method"`
	Count  int64  `json:"count"`
	Total  int64  `json:"total"`
}

// ManagerPaymentReport merangkum pembayaran untuk dashboard manager.
// Pendapatan hanya menghitung transaksi PAID; settlement dilaporkan terpisah.
type ManagerPaymentReport struct {
	RangeLabel      string           `json:"range_label"`
	From            string           `json:"from"`
	To              string           `json:"to"`
	TotalRevenue    int64            `json:"total_revenue"`
	TotalSuccessful int64            `json:"total_successful"`
	TotalPending    int64            `json:"total_pending"`
	TotalFailed     int64            `json:"total_failed"`
	TotalExpired    int64            `json:"total_expired"`
	TotalSettled    int64            `json:"total_settled"`
	TotalUnsettled  int64            `json:"total_unsettled"`
	ByMethod        []MethodTotal    `json:"by_method"`
	StatusCounts    map[string]int64 `json:"status_counts"`
}

// PaymentSettingsView adalah tampilan aman konfigurasi gateway: tanpa rahasia.
type PaymentSettingsView struct {
	Provider                string     `json:"provider"`
	Environment             string     `json:"environment"`
	Enabled                 bool       `json:"enabled"`
	Connected               bool       `json:"connected"`
	Configured              bool       `json:"configured"`
	MerchantName            string     `json:"merchant_name"`
	SettlementBank          string     `json:"settlement_bank"`
	SettlementAccountMasked string     `json:"settlement_account_masked"`
	SettlementStatus        string     `json:"settlement_status"`
	WebhookURL              string     `json:"webhook_url"`
	LastSyncedAt            *time.Time `json:"last_synced_at,omitempty"`
}
