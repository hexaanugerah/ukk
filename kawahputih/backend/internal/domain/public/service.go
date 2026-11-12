package public

import (
	"context"

	"github.com/KawahPutih/backend/internal/domain/tickettype"
	"github.com/KawahPutih/backend/pkg/apperror"
)

type Service struct {
	repo          Repository
	ticketService *tickettype.Service
}

func NewService(repo Repository, ticketService *tickettype.Service) *Service {
	return &Service{repo: repo, ticketService: ticketService}
}

func (s *Service) GetDestinations(ctx context.Context) ([]Destination, error) {
	return s.repo.ListDestinations(ctx, true)
}

func (s *Service) GetDestinationByID(ctx context.Context, id string) (*Destination, error) {
	d, err := s.repo.FindDestinationByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if d == nil {
		return nil, ErrNotFound
	}
	return d, nil
}

func (s *Service) GetDestinationBySlug(ctx context.Context, slug string) (*Destination, error) {
	d, err := s.repo.FindDestinationBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if d == nil {
		return nil, ErrNotFound
	}
	return d, nil
}

func (s *Service) GetFacilities(ctx context.Context, category string) ([]Facility, error) {
	return s.repo.ListFacilities(ctx, category, true)
}

func (s *Service) GetFAQs(ctx context.Context, category string) ([]FAQ, error) {
	return s.repo.ListFAQs(ctx, category, true)
}

func (s *Service) GetReviews(ctx context.Context, destinationID string, limit int) ([]Review, error) {
	return s.repo.ListReviews(ctx, destinationID, true, limit)
}

func (s *Service) GetContents(ctx context.Context, contentType string, page, limit int) ([]Content, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.ListContents(ctx, contentType, ContentStatusPublished, limit, offset)
}

func (s *Service) GetContentBySlug(ctx context.Context, slug string) (*Content, error) {
	c, err := s.repo.FindContentBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if c == nil {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *Service) GetContentByID(ctx context.Context, id string) (*Content, error) {
	c, err := s.repo.FindContentByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if c == nil {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *Service) GetPrices(ctx context.Context) ([]PriceInfoResponse, error) {
	tickets, err := s.ticketService.ListActive(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	sold, err := s.repo.ListReviews(ctx, "", false, 0)
	_ = sold
	result := make([]PriceInfoResponse, 0, len(tickets))
	for _, t := range tickets {
		result = append(result, PriceInfoResponse{
			TicketID:   t.ID,
			TicketName: t.Name,
			Category:   t.Category,
			Price:      t.Price,
			Capacity:   t.Capacity,
			Available:  t.Capacity,
			IsActive:   t.IsActive,
		})
	}
	return result, nil
}
