package verification

import (
	"context"
	"time"

	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/pkg/apperror"
)

type Service struct {
	repo          Repository
	ticketService *ticketing.Service
}

func NewService(repo Repository, ticketService *ticketing.Service) *Service {
	return &Service{repo: repo, ticketService: ticketService}
}

func (s *Service) Verify(ctx context.Context, req VerifyRequest, staffID, staffName string) (*VerifyResponse, error) {
	// Pakai service ticketing untuk memvalidasi tiket
	result, err := s.ticketService.Validate(ctx, req.TicketID, req.Gate, staffID, staffName)
	if err != nil {
		return nil, err
	}

	// Buat record verifikasi
	scannedAt := time.Now()
	if result.Ticket.CheckedInAt != nil {
		scannedAt = *result.Ticket.CheckedInAt
	}

	verification := &Verification{
		TicketID:    result.Ticket.ID,
		BookingID:   result.Ticket.BookingID,
		VisitorName: result.Ticket.VisitorName,
		Gate:        req.Gate,
		PetugasID:   staffID,
		PetugasName: staffName,
		Status:      result.Status,
		ScannedAt:   scannedAt,
		Note:        req.Note,
	}

	if err := s.repo.Create(ctx, verification); err != nil {
		return nil, apperror.Internal(err)
	}

	return &VerifyResponse{
		Success: result.Status == ticketing.ScanValid,
		Message: result.Message,
		Status:  result.Status,
		Ticket:  toVerificationTicketResponse(result.Ticket),
	}, nil
}

func toVerificationTicketResponse(t *ticketing.TicketResponse) *TicketResponse {
	return &TicketResponse{
		ID:            t.ID,
		BookingID:     t.BookingID,
		VisitorName:   t.VisitorName,
		Email:         t.Email,
		VisitDate:     t.VisitDate,
		TicketName:    t.TicketName,
		Quantity:      t.Quantity,
		Total:         t.Total,
		Status:        t.Status,
		GeneratedAt:   t.GeneratedAt,
		CheckedInAt:   t.CheckedInAt,
		CheckedInGate: t.CheckedInGate,
		QRData:        t.QRData,
	}
}

func (s *Service) GetVerification(ctx context.Context, id string) (*VerificationResponse, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if v == nil {
		return nil, ErrNotFound
	}
	return ToVerificationResponse(v), nil
}

func (s *Service) ListVerifications(ctx context.Context, page, limit int, gate, petugasID, from, to string) ([]VerificationResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	items, total, err := s.repo.List(ctx, offset, limit, gate, petugasID, from, to)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return ToVerificationResponseList(items), total, nil
}
