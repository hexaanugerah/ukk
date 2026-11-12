package visitorrule

import (
	"context"
)

type Usecase struct {
	svc *Service
}

func NewUsecase(svc *Service) *Usecase {
	return &Usecase{svc: svc}
}

func (u *Usecase) Create(ctx context.Context, req CreateRequest) (*VisitorRule, error) {
	return u.svc.Create(ctx, req)
}

func (u *Usecase) Update(ctx context.Context, id string, req UpdateRequest) (*VisitorRule, error) {
	return u.svc.Update(ctx, id, req)
}

func (u *Usecase) Delete(ctx context.Context, id string) error {
	return u.svc.Delete(ctx, id)
}

func (u *Usecase) Get(ctx context.Context, id string) (*VisitorRule, error) {
	return u.svc.Get(ctx, id)
}

func (u *Usecase) List(ctx context.Context, offset, limit int, category string, publishedOnly bool) ([]*VisitorRule, int64, error) {
	return u.svc.List(ctx, offset, limit, category, publishedOnly)
}

func (u *Usecase) ListPublished(ctx context.Context, category string) ([]*VisitorRule, error) {
	return u.svc.ListPublished(ctx, category)
}
