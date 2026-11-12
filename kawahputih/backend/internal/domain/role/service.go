package role

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

func (s *Service) ListRoles(ctx context.Context) ([]RoleResponse, error) {
	roles, err := s.repo.ListRoles(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return toRoleResponseList(roles), nil
}

func (s *Service) GetRoleNames(ctx context.Context, userID string) ([]string, error) {
	names, err := s.repo.GetRoleNamesForUser(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return names, nil
}

// AssignDefaultRole memberi pengguna baru (daftar sendiri atau via Google) role pengunjung. Idempoten: bila sudah punya role, tidak melakukan apa-apa, bukan error.
func (s *Service) AssignDefaultRole(ctx context.Context, userID string) error {
	hasRole, err := s.repo.UserHasAnyRole(ctx, userID)
	if err != nil {
		return apperror.Internal(err)
	}
	if hasRole {
		return nil
	}
	defaultRole, err := s.repo.FindRoleByName(ctx, DefaultRoleName)
	if err != nil {
		return apperror.Internal(err)
	}
	if defaultRole == nil {
		return ErrRoleNotFound
	}
	if err := s.repo.AssignRole(ctx, userID, defaultRole.ID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) ReplaceRoles(ctx context.Context, userID string, roleNames []string) error {
	roles, err := s.repo.FindRolesByNames(ctx, roleNames)
	if err != nil {
		return apperror.Internal(err)
	}
	roleIDs := make([]string, 0, len(roles))
	for _, r := range roles {
		roleIDs = append(roleIDs, r.ID)
	}
	if err := s.repo.ReplaceRoles(ctx, userID, roleIDs); err != nil {
		return apperror.Internal(err)
	}
	return nil
}
