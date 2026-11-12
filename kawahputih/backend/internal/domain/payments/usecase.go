package payments

import (
	"context"

	"github.com/KawahPutih/backend/pkg/response"
	"github.com/go-playground/validator/v10"
)

type Usecase struct {
	service *Service
}

func NewUsecase(service *Service) *Usecase { return &Usecase{service: service} }

var validate = validator.New()

func Validate(s interface{}) []response.FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var out []response.FieldError
	for _, fe := range err.(validator.ValidationErrors) {
		out = append(out, response.FieldError{Field: fe.Field(), Message: fe.Field() + " is invalid"})
	}
	return out
}

type ProcessRequest struct {
	BookingID string `json:"booking_id" validate:"required,len=36"`
	Forced    string `json:"forced" validate:"omitempty,oneof=SUCCESS FAILED EXPIRED"`
	// Method adalah metode pembayaran pilihan pengunjung. Opsional agar pemanggil lama (hanya QRIS) tetap jalan.
	Method string `json:"method" validate:"omitempty,oneof=QRIS GOPAY OVO DANA SHOPEEPAY BCA_VA BNI_VA MANDIRI_VA BRI_VA CARD ALFAMART INDOMARET"`
}

type SetStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=PENDING PROCESSING PAID FAILED EXPIRED REFUNDED"`
}

func (u *Usecase) Process(ctx context.Context, req ProcessRequest, requesterID string, privileged bool) (*ProcessResult, error) {
	resp, pay, outcome, err := u.service.Process(ctx, req.BookingID, requesterID, privileged, req.Forced, req.Method)
	if err != nil {
		return nil, err
	}
	return &ProcessResult{Booking: resp, Payment: toPaymentDTO(pay), Outcome: outcome}, nil
}

func (u *Usecase) Expire(ctx context.Context, bookingID, requesterID string, privileged bool) error {
	return u.service.Expire(ctx, bookingID, requesterID, privileged)
}

func (u *Usecase) GetDetail(ctx context.Context, bookingID, requesterID string, privileged bool) (*PaymentDetail, error) {
	return u.service.GetDetail(ctx, bookingID, requesterID, privileged)
}

func (u *Usecase) GetByID(ctx context.Context, id, requesterID string, privileged bool) (*PaymentDTO, error) {
	return u.service.GetByID(ctx, id, requesterID, privileged)
}

func (u *Usecase) MarkDownloaded(ctx context.Context, bookingID, requesterID string, privileged bool) error {
	return u.service.MarkDownloaded(ctx, bookingID, requesterID, privileged)
}

func (u *Usecase) ListAll(ctx context.Context, page, limit int, status string) ([]PaymentDTO, int64, error) {
	return u.service.ListAll(ctx, page, limit, status)
}

// Webhook menangani notifikasi pembayaran Midtrans.
func (u *Usecase) HandleMidtransWebhook(ctx context.Context, orderID, transactionStatus, fraudStatus, signatureKey, grossAmount string) error {
	return u.service.HandleMidtransNotification(ctx, orderID, transactionStatus, fraudStatus, signatureKey, grossAmount)
}

func (u *Usecase) SetStatusByAdmin(ctx context.Context, paymentID, status string) (*ProcessResult, error) {
	pay, resp, err := u.service.SetStatusByAdmin(ctx, paymentID, status)
	if err != nil {
		return nil, err
	}
	return &ProcessResult{Booking: resp, Payment: toPaymentDTO(pay), Outcome: status}, nil
}
