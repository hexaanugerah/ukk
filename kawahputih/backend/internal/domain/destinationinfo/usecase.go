package destinationinfo

import (
	"context"
)

type Usecase struct {
	svc *Service
}

func NewUsecase(svc *Service) *Usecase {
	return &Usecase{svc: svc}
}

func (u *Usecase) Create(ctx context.Context, authorID string, req CreateRequest) (*DestinationInfo, error) {
	return u.svc.Create(ctx, authorID, req)
}

func (u *Usecase) Update(ctx context.Context, id, actorID string, req UpdateRequest) (*DestinationInfo, error) {
	return u.svc.Update(ctx, id, actorID, req)
}

func (u *Usecase) Publish(ctx context.Context, id, actorID string) error {
	return u.svc.Publish(ctx, id, actorID)
}

func (u *Usecase) Archive(ctx context.Context, id, actorID string) error {
	return u.svc.Archive(ctx, id, actorID)
}

func (u *Usecase) Delete(ctx context.Context, id, actorID string) error {
	return u.svc.Delete(ctx, id, actorID)
}

func (u *Usecase) Get(ctx context.Context, id string) (*DestinationInfo, error) {
	return u.svc.Get(ctx, id)
}

func (u *Usecase) GetPublished(ctx context.Context) (*DestinationInfo, error) {
	return u.svc.GetPublished(ctx)
}

func (u *Usecase) List(ctx context.Context, offset, limit int, status string) ([]*DestinationInfo, int64, error) {
	return u.svc.List(ctx, offset, limit, status)
}
