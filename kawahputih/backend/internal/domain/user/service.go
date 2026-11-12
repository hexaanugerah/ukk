package user

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

func (s *Service) List(ctx context.Context, offset, limit int, search string) ([]AdminUser, int64, error) {
	users, total, err := s.repo.List(ctx, offset, limit, search)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return users, total, nil
}

func (s *Service) SetActive(ctx context.Context, id string, active bool, actorID string) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return apperror.Internal(err)
	}
	if u == nil {
		return ErrUserNotFound
	}
	if err := s.repo.SetActive(ctx, id, active, actorID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}
