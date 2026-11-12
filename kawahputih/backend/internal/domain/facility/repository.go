package facility

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, facility *Facility) error
	Update(ctx context.Context, facility *Facility) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Facility, error)
	List(ctx context.Context, offset, limit int, status string, publishedOnly bool) ([]*Facility, int64, error)
	ListPublished(ctx context.Context) ([]*Facility, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, facility *Facility) error {
	return r.db.WithContext(ctx).Create(facility).Error
}

func (r *repository) Update(ctx context.Context, facility *Facility) error {
	return r.db.WithContext(ctx).Save(facility).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Facility{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*Facility, error) {
	var facility Facility
	err := r.db.WithContext(ctx).First(&facility, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &facility, nil
}

func (r *repository) List(ctx context.Context, offset, limit int, status string, publishedOnly bool) ([]*Facility, int64, error) {
	q := r.db.WithContext(ctx).Model(&Facility{}).Order("status ASC, sort_order ASC, created_at DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if publishedOnly {
		q = q.Where("is_published = ?", true)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []*Facility
	if err := q.Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repository) ListPublished(ctx context.Context) ([]*Facility, error) {
	var items []*Facility
	err := r.db.WithContext(ctx).Where("is_published = ?", true).Order("sort_order ASC, created_at DESC").Find(&items).Error
	return items, err
}
