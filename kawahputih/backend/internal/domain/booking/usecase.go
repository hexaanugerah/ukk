package booking

import (
	"context"
	"time"

	"github.com/KawahPutih/backend/pkg/response"
	"github.com/go-playground/validator/v10"
)

// Usecase menghubungkan handler ke service dan memegang ketergantungan lintas domain untuk harga (TicketPricer dari domain tickettype).
type Usecase struct {
	service *Service
	pricer  TicketPricer
}

func NewUsecase(service *Service, pricer TicketPricer) *Usecase {
	return &Usecase{service: service, pricer: pricer}
}

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

func (u *Usecase) CreateBooking(ctx context.Context, userID string, req CreateRequest) (*CreateBookingResult, error) {
	return u.service.Create(ctx, userID, u.pricer, req)
}

func (u *Usecase) Get(ctx context.Context, id, requesterID string, privileged bool, reader PaymentReader) (*BookingResponse, error) {
	return u.service.Get(ctx, id, requesterID, privileged, reader)
}

func (u *Usecase) GetByCode(ctx context.Context, code, requesterID string, privileged bool, reader PaymentReader) (*BookingResponse, error) {
	return u.service.GetByCode(ctx, code, requesterID, privileged, reader)
}

func (u *Usecase) ListMine(ctx context.Context, userID string, page, limit int, status string, reader PaymentReader) ([]BookingResponse, int64, error) {
	return u.service.ListMine(ctx, userID, page, limit, status, reader)
}

func (u *Usecase) ListForUser(ctx context.Context, userID string, page, limit int, status string, reader PaymentReader) ([]BookingResponse, int64, error) {
	return u.service.ListForUser(ctx, userID, page, limit, status, reader)
}

func (u *Usecase) ListAll(ctx context.Context, page, limit int, status, search string, reader PaymentReader) ([]BookingResponse, int64, error) {
	return u.service.ListAll(ctx, page, limit, status, search, reader)
}

func (u *Usecase) ListToday(ctx context.Context, date string, page, limit int, reader PaymentReader) ([]BookingResponse, int64, error) {
	return u.service.ListByVisitDate(ctx, date, page, limit, reader)
}

func (u *Usecase) Cancel(ctx context.Context, id, requesterID string, privileged bool) error {
	return u.service.Cancel(ctx, id, requesterID, privileged)
}

func (u *Usecase) SetStatus(ctx context.Context, id, status string) (*BookingResponse, error) {
	return u.service.SetStatus(ctx, id, status)
}

func (u *Usecase) GetSnapshot(ctx context.Context, bookingID, requesterID string, privileged bool) (*Snapshot, error) {
	return u.service.GetSnapshot(ctx, bookingID, requesterID, privileged)
}

// SetCheckedIn dipakai domain scanner (lewat adapter bootstrap) untuk menandai booking terpakai setelah scan gerbang berhasil.
func (u *Usecase) SetCheckedIn(ctx context.Context, bookingID, gate string, at time.Time) error {
	return u.service.SetCheckedIn(ctx, bookingID, gate, at)
}

// MarkPaid dipicu domain payments setelah konfirmasi pembayaran berhasil.
func (u *Usecase) MarkPaid(ctx context.Context, bookingID string) error {
	return u.service.MarkPaid(ctx, bookingID)
}

// MarkExpired dipicu domain payments (expire/sweeper).
func (u *Usecase) MarkExpired(ctx context.Context, bookingID string) error {
	return u.service.MarkExpired(ctx, bookingID)
}

// GetResponse memuat respons booking lengkap untuk konsumen lintas domain (domain payments menyematkannya).
func (u *Usecase) GetResponse(ctx context.Context, bookingID, requesterID string, privileged bool) (*BookingResponse, error) {
	return u.service.Get(ctx, bookingID, requesterID, privileged, nil)
}

// StalePendingIDs membuka query sweeper (digerakkan domain payments).
func (u *Usecase) StalePendingIDs(ctx context.Context, cutoff time.Time, limit int) ([]string, error) {
	return u.service.StalePendingIDs(ctx, cutoff, limit)
}
