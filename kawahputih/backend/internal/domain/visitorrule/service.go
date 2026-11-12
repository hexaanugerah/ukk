package visitorrule

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

func (s *Service) Create(ctx context.Context, req CreateRequest) (*VisitorRule, error) {
	rule := &VisitorRule{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Priority:    req.Priority,
		SortOrder:   req.SortOrder,
		IsPublished: req.IsPublished,
	}
	if rule.Priority == "" {
		rule.Priority = PriorityNormal
	}
	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, apperror.Internal(err)
	}
	return rule, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*VisitorRule, error) {
	rule, err := s.mustFind(ctx, id)
	if err != nil {
		return nil, err
	}
	rule.Title = req.Title
	rule.Description = req.Description
	rule.Category = req.Category
	rule.Priority = req.Priority
	rule.SortOrder = req.SortOrder
	rule.IsPublished = req.IsPublished
	if rule.Priority == "" {
		rule.Priority = PriorityNormal
	}
	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, apperror.Internal(err)
	}
	return rule, nil
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

func (s *Service) Get(ctx context.Context, id string) (*VisitorRule, error) {
	return s.mustFind(ctx, id)
}

func (s *Service) List(ctx context.Context, offset, limit int, category string, publishedOnly bool) ([]*VisitorRule, int64, error) {
	items, total, err := s.repo.List(ctx, offset, limit, category, publishedOnly)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return items, total, nil
}

func (s *Service) ListPublished(ctx context.Context, category string) ([]*VisitorRule, error) {
	return s.repo.ListPublished(ctx, category)
}

func (s *Service) mustFind(ctx context.Context, id string) (*VisitorRule, error) {
	rule, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if rule == nil {
		return nil, ErrNotFound
	}
	return rule, nil
}
