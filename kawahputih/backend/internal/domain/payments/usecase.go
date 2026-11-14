package payments

import (
	"context"
)

type Usecase struct {
	service *Service
}

func NewUsecase(service *Service) *Usecase {
	return &Usecase{service: service}
}

func (u *Usecase) Process(ctx context.Context, req ProcessRequest, userID string, privileged bool) (*ProcessResult, error) {
	if req.BookingID == "" {
		return nil, ErrNotFound
	}
	if req.Method != "" && !ValidMethod(req.Method) {
		return nil, ErrInvalidMethod
	}

	bookingResp, pay, outcome, err := u.service.Process(ctx, req.BookingID, userID, privileged, req.Forced, req.Method, req.EWalletNumber)
	if err != nil {
		return nil, err
	}
	return &ProcessResult{
		Booking: bookingResp,
		Payment: PaymentDTO{
			ID:            pay.ID,
			BookingID:     pay.BookingID,
			Method:        pay.Method,
			EWalletNumber: pay.EWalletNumber,
			Amount:        pay.Amount,
			Status:        pay.Status,
			CreatedAt:     pay.CreatedAt,
			PaidAt:        pay.PaidAt,
			ExpiredAt:     pay.ExpiredAt,
		},
		Outcome: outcome,
	}, nil
}

func (u *Usecase) Expire(ctx context.Context, bookingID, userID string, privileged bool) error {
	return u.service.Expire(ctx, bookingID, userID, privileged)
}

func (u *Usecase) GetDetail(ctx context.Context, bookingID, userID string, privileged bool) (*PaymentDetail, error) {
	return u.service.GetDetail(ctx, bookingID, userID, privileged)
}

func (u *Usecase) MarkDownloaded(ctx context.Context, bookingID, userID string, privileged bool) error {
	return u.service.MarkDownloaded(ctx, bookingID, userID, privileged)
}

func (u *Usecase) ListAll(ctx context.Context, page, limit int, status string) ([]PaymentDTO, int64, error) {
	return u.service.ListAll(ctx, page, limit, status)
}

func (u *Usecase) SetStatusByAdmin(ctx context.Context, paymentID, status string) (*ProcessResult, error) {
	pay, bookingResp, err := u.service.SetStatusByAdmin(ctx, paymentID, status)
	if err != nil {
		return nil, err
	}
	return &ProcessResult{
		Booking: bookingResp,
		Payment: PaymentDTO{
			ID:            pay.ID,
			BookingID:     pay.BookingID,
			Method:        pay.Method,
			EWalletNumber: pay.EWalletNumber,
			Amount:        pay.Amount,
			Status:        pay.Status,
			CreatedAt:     pay.CreatedAt,
			PaidAt:        pay.PaidAt,
			ExpiredAt:     pay.ExpiredAt,
		},
		Outcome: "UPDATED",
	}, nil
}

func (u *Usecase) GetByID(ctx context.Context, id, userID string, privileged bool) (*PaymentDTO, error) {
	return u.service.GetByID(ctx, id, userID, privileged)
}

func (u *Usecase) HandleMidtransWebhook(ctx context.Context, notificationID, orderID, transactionStatus, statusCode, fraudStatus, signatureKey, grossAmount string) error {
	return u.service.HandleMidtransNotification(ctx, notificationID, orderID, transactionStatus, statusCode, fraudStatus, signatureKey, grossAmount)
}

func (u *Usecase) ListFiltered(ctx context.Context, page, limit int, filter ListFilter) ([]PaymentDTO, int64, error) {
	return u.service.ListFiltered(ctx, page, limit, filter)
}

func (u *Usecase) AttachProof(ctx context.Context, bookingID, userID string, privileged bool, proofURL string) (*PaymentDTO, error) {
	return u.service.AttachProof(ctx, bookingID, userID, privileged, proofURL)
}

func (u *Usecase) ManagerReport(ctx context.Context, from, to string) (*ManagerPaymentReport, error) {
	return u.service.ManagerReport(ctx, from, to)
}

func (u *Usecase) SetEWalletNumber(ctx context.Context, paymentID, ewalletNumber string) (*PaymentDTO, error) {
	pay, err := u.service.SetEWalletNumber(ctx, paymentID, ewalletNumber)
	if err != nil {
		return nil, err
	}
	return &PaymentDTO{
		ID:            pay.ID,
		BookingID:     pay.BookingID,
		Method:        pay.Method,
		EWalletNumber: pay.EWalletNumber,
		Amount:        pay.Amount,
		Status:        pay.Status,
		CreatedAt:     pay.CreatedAt,
		PaidAt:        pay.PaidAt,
		ExpiredAt:     pay.ExpiredAt,
	}, nil
}
