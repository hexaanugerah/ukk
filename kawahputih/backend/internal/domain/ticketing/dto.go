package ticketing

import (
	"context"
	"time"

	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/go-playground/validator/v10"
)

// TicketResponse cocok persis dengan tipe ETicket frontend.
type TicketResponse struct {
	ID            string     `json:"id"`
	BookingID     string     `json:"booking_id"`
	BookingCode   string     `json:"booking_code"`
	VisitorName   string     `json:"visitor_name"`
	Email         string     `json:"email"`
	VisitDate     string     `json:"visit_date"`
	TicketName    string     `json:"ticket_name"`
	Quantity      int        `json:"quantity"`
	Total         int64      `json:"total"`
	Status        string     `json:"status"`
	GeneratedAt   time.Time  `json:"generated_at"`
	CheckedInAt   *time.Time `json:"checked_in_at,omitempty"`
	CheckedInGate string     `json:"checked_in_gate,omitempty"`
	QRData        string     `json:"qr_data"`
}

func toTicketResponse(t *Ticket) *TicketResponse {
	return &TicketResponse{
		ID: t.ID, BookingID: t.BookingID, BookingCode: t.BookingCode, VisitorName: t.VisitorName, Email: t.Email,
		VisitDate: t.VisitDate, TicketName: t.TicketName, Quantity: t.Quantity,
		Total: t.Total, Status: t.Status, GeneratedAt: t.GeneratedAt,
		CheckedInAt: t.CheckedInAt, CheckedInGate: t.CheckedInGate, QRData: t.QRData,
	}
}

var (
	ErrForbidden = apperror.Forbidden("Anda tidak memiliki akses.")
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

type ValidateRequest struct {
	TicketID string `json:"ticket_id" validate:"required,min=4,max=120"`
	Gate     string `json:"gate" validate:"omitempty,max=100"`
	Petugas  string `json:"petugas" validate:"omitempty,max=150"`
}

type ValidateResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Status  string          `json:"status"`
	Ticket  *TicketResponse `json:"ticket,omitempty"`
}

func (u *Usecase) ValidateTicket(ctx context.Context, req ValidateRequest, staffID, staffName string) (*ValidateResponse, error) {
	result, err := u.service.Validate(ctx, req.TicketID, req.Gate, staffID, staffName)
	if err != nil {
		return nil, err
	}
	return &ValidateResponse{
		Success: result.Status == ScanValid,
		Message: result.Message,
		Status:  result.Status,
		Ticket:  result.Ticket,
	}, nil
}

// ErrTicketNotFound: pencarian hanya-baca tidak menemukan apa pun untuk referensi itu.
var ErrTicketNotFound = apperror.NotFound("Data tiket tidak ditemukan.")

func (u *Usecase) LookupTicket(ctx context.Context, ref string) (*TicketResponse, error) {
	t, err := u.service.GetByIDOrQR(ctx, ref)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTicketNotFound
	}
	return toTicketResponse(t), nil
}

// ListMyTickets mengembalikan e-tiket milik pengunjung yang sedang login.
func (u *Usecase) ListMyTickets(ctx context.Context, userID string) ([]TicketResponse, error) {
	return u.service.ListMine(ctx, userID)
}

// CanAccess memeriksa kepemilikan e-tiket sebelum handler mengembalikan datanya.
func (u *Usecase) CanAccess(ctx context.Context, bookingID, userID string, privileged bool) (bool, error) {
	return u.service.CanAccess(ctx, bookingID, userID, privileged)
}

func (u *Usecase) GetTicket(ctx context.Context, id string) (*TicketResponse, error) {
	t, err := u.service.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTicketNotFound
	}
	return t, nil
}

func (u *Usecase) GetTicketByCode(ctx context.Context, code string) (*TicketResponse, error) {
	t, err := u.service.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTicketNotFound
	}
	return t, nil
}

func (u *Usecase) GetTicketByBooking(ctx context.Context, bookingID string) (*TicketResponse, error) {
	t, err := u.service.GetByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTicketNotFound
	}
	return t, nil
}

func (u *Usecase) History(ctx context.Context, page, limit int, gate, petugasID, from, to string) ([]ScanRecord, int64, error) {
	return u.service.History(ctx, page, limit, gate, petugasID, from, to)
}

func (u *Usecase) Stats(ctx context.Context) (map[string]int64, error) {
	return u.service.Stats(ctx)
}

func (u *Usecase) Monitoring(ctx context.Context, date string) (*MonitoringResponse, error) {
	return u.service.Monitoring(ctx, date)
}
