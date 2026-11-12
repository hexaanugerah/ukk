package article

import (
	"context"
	"strconv"
	"time"

	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/slugutil"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, authorID string, req CreateRequest) (*Article, error) {
	slug, err := s.uniqueSlug(ctx, req.Title, "")
	if err != nil {
		return nil, err
	}
	a := &Article{
		Title: req.Title, Slug: slug, Excerpt: req.Excerpt, Content: req.Content,
		CoverImage: req.CoverImage, Status: StatusDraft, AuthorID: authorID,
		CreatedBy: authorID, UpdatedBy: authorID,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, apperror.Internal(err)
	}
	return a, nil
}

func (s *Service) Update(ctx context.Context, id, actorID string, req UpdateRequest) (*Article, error) {
	a, err := s.mustFind(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Title != req.Title {
		slug, err := s.uniqueSlug(ctx, req.Title, a.ID)
		if err != nil {
			return nil, err
		}
		a.Slug = slug
	}
	a.Title, a.Excerpt, a.Content, a.CoverImage = req.Title, req.Excerpt, req.Content, req.CoverImage
	a.UpdatedBy = actorID
	if err := s.repo.Update(ctx, a); err != nil {
		return nil, apperror.Internal(err)
	}
	return a, nil
}

func (s *Service) Publish(ctx context.Context, id, actorID string) error {
	a, err := s.mustFind(ctx, id)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	a.Status, a.PublishedAt, a.UpdatedBy = StatusPublished, &now, actorID
	if err := s.repo.Update(ctx, a); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) Archive(ctx context.Context, id, actorID string) error {
	a, err := s.mustFind(ctx, id)
	if err != nil {
		return err
	}
	a.Status, a.UpdatedBy = StatusArchived, actorID
	if err := s.repo.Update(ctx, a); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, id, actorID string) error {
	a, err := s.mustFind(ctx, id)
	if err != nil {
		return err
	}
	// deleted_by dicatat SEBELUM soft delete karena soft delete GORM hanya mengisi deleted_at. Urutan UPDATE lalu DELETE inilah yang membuat deleted_by tetap ada untuk query audit.
	a.DeletedBy = &actorID
	if err := s.repo.Update(ctx, a); err != nil {
		return apperror.Internal(err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id string) (*Article, error) { return s.mustFind(ctx, id) }

func (s *Service) List(ctx context.Context, offset, limit int, status string) ([]Article, int64, error) {
	articles, total, err := s.repo.List(ctx, offset, limit, status)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return articles, total, nil
}

func (s *Service) mustFind(ctx context.Context, id string) (*Article, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if a == nil {
		return nil, ErrArticleNotFound
	}
	return a, nil
}

func (s *Service) uniqueSlug(ctx context.Context, title, excludeID string) (string, error) {
	base := slugutil.Generate(title)
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
