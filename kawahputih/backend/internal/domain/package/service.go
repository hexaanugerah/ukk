package tourismpackage

import (
	"context"
	"strconv"

	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/slugutil"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, actorID string, req CreateRequest) (*Package, error) {
	slug, err := s.uniqueSlug(ctx, req.Name, "")
	if err != nil {
		return nil, err
	}
	p := &Package{
		Name: req.Name, Slug: slug, Description: req.Description, CoverImage: req.CoverImage,
		PriceCents: req.PriceCents, Currency: "IDR", DurationHours: req.DurationHours,
		MaxCapacity: req.MaxCapacity, IsActive: true,
		CreatedBy: actorID, UpdatedBy: actorID,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, apperror.Internal(err)
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id, actorID string, req UpdateRequest) (*Package, error) {
	p, err := s.mustFind(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Name != req.Name {
		slug, err := s.uniqueSlug(ctx, req.Name, p.ID)
		if err != nil {
			return nil, err
		}
		p.Slug = slug
	}
	p.Name, p.Description, p.CoverImage = req.Name, req.Description, req.CoverImage
	p.PriceCents, p.DurationHours, p.MaxCapacity, p.IsActive = req.PriceCents, req.DurationHours, req.MaxCapacity, req.IsActive
	p.UpdatedBy = actorID
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.Internal(err)
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, id, actorID string) error {
	p, err := s.mustFind(ctx, id)
	if err != nil {
		return err
	}
	p.DeletedBy = &actorID
	if err := s.repo.Update(ctx, p); err != nil {
		return apperror.Internal(err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id string) (*Package, error) { return s.mustFind(ctx, id) }

func (s *Service) List(ctx context.Context, offset, limit int, activeOnly bool) ([]Package, int64, error) {
	packages, total, err := s.repo.List(ctx, offset, limit, activeOnly)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return packages, total, nil
}

func (s *Service) mustFind(ctx context.Context, id string) (*Package, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if p == nil {
		return nil, ErrPackageNotFound
	}
	return p, nil
}

func (s *Service) uniqueSlug(ctx context.Context, name, excludeID string) (string, error) {
	base := slugutil.Generate(name)
	slug := base
	for i := 2; ; i++ {
		exists, err := s.repo.SlugExists(ctx, slug, excludeID)
		if err != nil {
			return "", apperror.Internal(err)
		}
		if !exists {
			return slug, nil
		}
		slug = base + "-" + strconv.Itoa(i)
	}
}
