package gallery

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, item *Item) error
	Update(ctx context.Context, item *Item) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Item, error)
	List(ctx context.Context, offset, limit int, category string) ([]Item, int64, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, item *Item) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *repository) Update(ctx context.Context, item *Item) error {
	return r.db.WithContext(ctx).Save(item).Error
}
func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Item{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*Item, error) {
	var item Item
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) List(ctx context.Context, offset, limit int, category string) ([]Item, int64, error) {
	q := r.db.WithContext(ctx).Model(&Item{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []Item
	if err := q.Order("sort_order ASC, created_at DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
