package visitorrule

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, rule *VisitorRule) error
	Update(ctx context.Context, rule *VisitorRule) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*VisitorRule, error)
	List(ctx context.Context, offset, limit int, category string, publishedOnly bool) ([]*VisitorRule, int64, error)
	ListPublished(ctx context.Context, category string) ([]*VisitorRule, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, rule *VisitorRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *repository) Update(ctx context.Context, rule *VisitorRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&VisitorRule{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*VisitorRule, error) {
	var rule VisitorRule
	err := r.db.WithContext(ctx).First(&rule, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *repository) List(ctx context.Context, offset, limit int, category string, publishedOnly bool) ([]*VisitorRule, int64, error) {
	q := r.db.WithContext(ctx).Model(&VisitorRule{}).Order("category ASC, priority DESC, sort_order ASC, created_at DESC")
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
	var items []*VisitorRule
	if err := q.Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repository) ListPublished(ctx context.Context, category string) ([]*VisitorRule, error) {
	q := r.db.WithContext(ctx).Where("is_published = ?", true).Order("category ASC, priority DESC, sort_order ASC, created_at DESC")
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var items []*VisitorRule
	err := q.Find(&items).Error
	return items, err
}
