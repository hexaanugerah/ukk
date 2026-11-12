package article

import "context"

type Usecase struct {
	service *Service
}

func NewUsecase(service *Service) *Usecase {
	return &Usecase{service: service}
}

func (u *Usecase) Create(ctx context.Context, authorID string, req CreateRequest) (Response, error) {
	a, err := u.service.Create(ctx, authorID, req)
	if err != nil {
		return Response{}, err
	}
	return toResponse(a), nil
}

func (u *Usecase) Update(ctx context.Context, id, actorID string, req UpdateRequest) (Response, error) {
	a, err := u.service.Update(ctx, id, actorID, req)
	if err != nil {
		return Response{}, err
	}
	return toResponse(a), nil
}

func (u *Usecase) Publish(ctx context.Context, id, actorID string) error {
	return u.service.Publish(ctx, id, actorID)
}
func (u *Usecase) Archive(ctx context.Context, id, actorID string) error {
	return u.service.Archive(ctx, id, actorID)
}
func (u *Usecase) Delete(ctx context.Context, id, actorID string) error {
	return u.service.Delete(ctx, id, actorID)
}

func (u *Usecase) Get(ctx context.Context, id string) (Response, error) {
	a, err := u.service.Get(ctx, id)
	if err != nil {
		return Response{}, err
	}
	return toResponse(a), nil
}

func (u *Usecase) List(ctx context.Context, page, perPage int, status string) ([]Response, int64, error) {
	page, perPage = normalizePaging(page, perPage)
	articles, total, err := u.service.List(ctx, (page-1)*perPage, perPage, status)
	if err != nil {
		return nil, 0, err
	}
	return toResponseList(articles), total, nil
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
