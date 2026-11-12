package faq

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, faq *FAQ) error
	Update(ctx context.Context, faq *FAQ) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*FAQ, error)
	List(ctx context.Context, offset, limit int, category string, publishedOnly bool) ([]*FAQ, int64, error)
	ListPublished(ctx context.Context, category string) ([]*FAQ, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, faq *FAQ) error {
	return r.db.WithContext(ctx).Create(faq).Error
}

func (r *repository) Update(ctx context.Context, faq *FAQ) error {
	return r.db.WithContext(ctx).Save(faq).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&FAQ{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*FAQ, error) {
	var faq FAQ
	err := r.db.WithContext(ctx).First(&faq, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &faq, nil
}

func (r *repository) List(ctx context.Context, offset, limit int, category string, publishedOnly bool) ([]*FAQ, int64, error) {
	q := r.db.WithContext(ctx).Model(&FAQ{}).Order("category ASC, sort_order ASC, created_at DESC")
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if publishedOnly {
		q = q.Where("is_published = ?", true)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []*FAQ
	if err := q.Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repository) ListPublished(ctx context.Context, category string) ([]*FAQ, error) {
	q := r.db.WithContext(ctx).Model(&FAQ{}).Where("is_published = ?", true).Order("category ASC, sort_order ASC, created_at DESC")
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var items []*FAQ
	err := q.Find(&items).Error
	return items, err
}
