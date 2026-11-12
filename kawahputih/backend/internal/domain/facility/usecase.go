package facility

import (
	"context"
)

type Usecase struct {
	svc *Service
}

func NewUsecase(svc *Service) *Usecase {
	return &Usecase{svc: svc}
}

func (u *Usecase) Create(ctx context.Context, req CreateRequest) (*Facility, error) {
	return u.svc.Create(ctx, req)
}

func (u *Usecase) Update(ctx context.Context, id string, req UpdateRequest) (*Facility, error) {
	return u.svc.Update(ctx, id, req)
}

func (u *Usecase) Delete(ctx context.Context, id string) error {
	return u.svc.Delete(ctx, id)
}

func (u *Usecase) Get(ctx context.Context, id string) (*Facility, error) {
	return u.svc.Get(ctx, id)
}

func (u *Usecase) List(ctx context.Context, offset, limit int, status string, publishedOnly bool) ([]*Facility, int64, error) {
	return u.svc.List(ctx, offset, limit, status, publishedOnly)
}

func (u *Usecase) ListPublished(ctx context.Context) ([]*Facility, error) {
	return u.svc.ListPublished(ctx)
}
