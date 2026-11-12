package article

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, a *Article) error
	Update(ctx context.Context, a *Article) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Article, error)
	FindPublishedBySlug(ctx context.Context, slug string) (*Article, error)
	List(ctx context.Context, offset, limit int, status string) ([]Article, int64, error)
	SlugExists(ctx context.Context, slug, excludeID string) (bool, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, a *Article) error {
	return r.db.WithContext(ctx).Create(a).Error
}
func (r *repository) Update(ctx context.Context, a *Article) error {
	return r.db.WithContext(ctx).Save(a).Error
}
func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Article{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*Article, error) {
	var a Article
	err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repository) FindPublishedBySlug(ctx context.Context, slug string) (*Article, error) {
	var a Article
	err := r.db.WithContext(ctx).First(&a, "slug = ? AND status = ?", slug, StatusPublished).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repository) List(ctx context.Context, offset, limit int, status string) ([]Article, int64, error) {
	q := r.db.WithContext(ctx).Model(&Article{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var articles []Article
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

func (r *repository) SlugExists(ctx context.Context, slug, excludeID string) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&Article{}).Where("slug = ?", slug)
	if excludeID != "" {
		q = q.Where("id <> ?", excludeID)
	}
	err := q.Count(&count).Error
	return count > 0, err
}
