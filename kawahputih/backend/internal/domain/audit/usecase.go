package audit

import (
	"context"
)

type Usecase struct {
	service *Service
}

func NewUsecase(service *Service) *Usecase {
	return &Usecase{service: service}
}

func (u *Usecase) List(ctx context.Context, req ListRequest) ([]AuditLogResponse, int64, error) {
	logs, total, err := u.service.List(ctx, req.Page, req.Limit, req.ActorID, req.Action, req.TargetType, req.TargetID, req.Status, req.From, req.To)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		resp = append(resp, toAuditLogResponse(&log))
	}
	return resp, total, nil
}

func (u *Usecase) GetByID(ctx context.Context, id string) (*AuditLogResponse, error) {
	log, err := u.service.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if log == nil {
		return nil, ErrNotFound
	}
	resp := toAuditLogResponse(log)
	return &resp, nil
}

func (u *Usecase) StatsByAction(ctx context.Context, from, to string) (map[string]int64, error) {
	return u.service.StatsByAction(ctx, from, to)
}

func (u *Usecase) StatsByActor(ctx context.Context, from, to string) (map[string]int64, error) {
	return u.service.StatsByActor(ctx, from, to)
}
