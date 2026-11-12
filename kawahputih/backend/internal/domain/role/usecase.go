package role

import "context"

// Usecase di sini tipis karena manajemen role tidak punya domain lain untuk dikoordinasi, jadi UseCase dan Service 1:1. Tetap dipisah demi konsistensi struktur dengan domain lain.
type Usecase struct {
	service *Service
}

func NewUsecase(service *Service) *Usecase {
	return &Usecase{service: service}
}

func (u *Usecase) ListRoles(ctx context.Context) ([]RoleResponse, error) {
	return u.service.ListRoles(ctx)
}

func (u *Usecase) AssignRoles(ctx context.Context, userID string, roleNames []string) error {
	return u.service.ReplaceRoles(ctx, userID, roleNames)
}
