package faq

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

func (s *Service) Create(ctx context.Context, req CreateRequest) (*FAQ, error) {
	faq := &FAQ{
		Question:    req.Question,
		Answer:      req.Answer,
		Category:    req.Category,
		SortOrder:   req.SortOrder,
		IsPublished: req.IsPublished,
	}
	if err := s.repo.Create(ctx, faq); err != nil {
		return nil, apperror.Internal(err)
	}
	return faq, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*FAQ, error) {
	faq, err := s.mustFind(ctx, id)
	if err != nil {
		return nil, err
	}
	faq.Question = req.Question
	faq.Answer = req.Answer
	faq.Category = req.Category
	faq.SortOrder = req.SortOrder
	faq.IsPublished = req.IsPublished
	if err := s.repo.Update(ctx, faq); err != nil {
		return nil, apperror.Internal(err)
	}
	return faq, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	_, err := s.mustFind(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id string) (*FAQ, error) { return s.mustFind(ctx, id) }

func (s *Service) List(ctx context.Context, offset, limit int, category string, publishedOnly bool) ([]*FAQ, int64, error) {
	items, total, err := s.repo.List(ctx, offset, limit, category, publishedOnly)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return items, total, nil
}

func (s *Service) ListPublished(ctx context.Context, category string) ([]*FAQ, error) {
	return s.repo.ListPublished(ctx, category)
}

func (s *Service) mustFind(ctx context.Context, id string) (*FAQ, error) {
	faq, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if faq == nil {
		return nil, ErrNotFound
	}
	return faq, nil
}
