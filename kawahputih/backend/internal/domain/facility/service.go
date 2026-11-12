package facility

import (
	"context"

	"github.com/KawahPutih/backend/pkg/apperror"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Facility, error) {
	facility := &Facility{
		Name:           req.Name,
		Description:    req.Description,
		Icon:           req.Icon,
		ImageURL:       req.ImageURL,
		Location:       req.Location,
		OperatingHours: req.OperatingHours,
		Status:         req.Status,
		IsPublished:    req.IsPublished,
		Category:       req.Category,
		SortOrder:      req.SortOrder,
	}
	if facility.Status == "" {
		facility.Status = StatusActive
	}
	if err := s.repo.Create(ctx, facility); err != nil {
		return nil, apperror.Internal(err)
	}
	return facility, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Facility, error) {
	facility, err := s.mustFind(ctx, id)
	if err != nil {
		return nil, err
	}
	facility.Name = req.Name
	facility.Description = req.Description
	facility.Icon = req.Icon
	facility.ImageURL = req.ImageURL
	facility.Location = req.Location
	facility.OperatingHours = req.OperatingHours
	facility.Status = req.Status
	facility.IsPublished = req.IsPublished
	facility.Category = req.Category
	facility.SortOrder = req.SortOrder
	if facility.Status == "" {
		facility.Status = StatusActive
	}
	if err := s.repo.Update(ctx, facility); err != nil {
		return nil, apperror.Internal(err)
	}
	return facility, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.mustFind(ctx, id); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id string) (*Facility, error) { return s.mustFind(ctx, id) }

func (s *Service) List(ctx context.Context, offset, limit int, status string, publishedOnly bool) ([]*Facility, int64, error) {
	items, total, err := s.repo.List(ctx, offset, limit, status, publishedOnly)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return items, total, nil
}

func (s *Service) ListPublished(ctx context.Context) ([]*Facility, error) {
	return s.repo.ListPublished(ctx)
}

func (s *Service) mustFind(ctx context.Context, id string) (*Facility, error) {
	facility, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if facility == nil {
		return nil, ErrNotFound
	}
	return facility, nil
}
