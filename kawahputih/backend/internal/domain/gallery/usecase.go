package gallery

import "context"

type Usecase struct {
	service *Service
}

func NewUsecase(service *Service) *Usecase {
	return &Usecase{service: service}
}

func (u *Usecase) Create(ctx context.Context, uploaderID string, req CreateRequest, image *ImageUpload) (Response, error) {
	item, err := u.service.Create(ctx, uploaderID, req, image)
	if err != nil {
		return Response{}, err
	}
	return toResponse(item), nil
}

func (u *Usecase) Update(ctx context.Context, id, actorID string, req UpdateRequest, image *ImageUpload) (Response, error) {
	item, err := u.service.Update(ctx, id, actorID, req, image)
	if err != nil {
		return Response{}, err
	}
	return toResponse(item), nil
}

func (u *Usecase) Get(ctx context.Context, id string) (Response, error) {
	item, err := u.service.Get(ctx, id)
	if err != nil {
		return Response{}, err
	}
	return toResponse(item), nil
}

func (u *Usecase) Delete(ctx context.Context, id, actorID string) error {
	return u.service.Delete(ctx, id, actorID)
}

func (u *Usecase) List(ctx context.Context, page, perPage int, category string) ([]Response, int64, error) {
	page, perPage = normalizePaging(page, perPage)
	items, total, err := u.service.List(ctx, (page-1)*perPage, perPage, category)
	if err != nil {
		return nil, 0, err
	}
	return toResponseList(items), total, nil
}

func (u *Usecase) Search(ctx context.Context, query string, page, perPage int) ([]Response, int64, error) {
	page, perPage = normalizePaging(page, perPage)
	items, total, err := u.service.Search(ctx, query, (page-1)*perPage, perPage)
	if err != nil {
		return nil, 0, err
	}
	return toResponseList(items), total, nil
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
