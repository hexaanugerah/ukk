package user

import "context"

// RoleLister adalah cara domain ini menjangkau role TANPA mengimpornya, pola yang sama seperti auth.RoleProvider. role.Service memenuhi interface ini secara struktural.
type RoleLister interface {
	GetRoleNames(ctx context.Context, userID string) ([]string, error)
}

type Usecase struct {
	service *Service
	roles   RoleLister
}

func NewUsecase(service *Service, roles RoleLister) *Usecase {
	return &Usecase{service: service, roles: roles}
}

func (u *Usecase) List(ctx context.Context, page, perPage int, search string) ([]AdminUserResponse, int64, error) {
	page, perPage = normalizePaging(page, perPage)
	users, total, err := u.service.List(ctx, (page-1)*perPage, perPage, search)
	if err != nil {
		return nil, 0, err
	}
	out := make([]AdminUserResponse, 0, len(users))
	for _, usr := range users {
		roleNames, err := u.roles.GetRoleNames(ctx, usr.ID)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, toResponse(usr, roleNames))
	}
	return out, total, nil
}

func (u *Usecase) SetActive(ctx context.Context, id string, active bool, actorID string) error {
	return u.service.SetActive(ctx, id, active, actorID)
}

func normalizePaging(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}
