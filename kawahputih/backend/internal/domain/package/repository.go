package tourismpackage

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, p *Package) error
	Update(ctx context.Context, p *Package) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Package, error)
	List(ctx context.Context, offset, limit int, activeOnly bool) ([]Package, int64, error)
	SlugExists(ctx context.Context, slug, excludeID string) (bool, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, p *Package) error {
	return r.db.WithContext(ctx).Create(p).Error
}
func (r *repository) Update(ctx context.Context, p *Package) error {
	return r.db.WithContext(ctx).Save(p).Error
}
func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Package{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*Package, error) {
	var p Package
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repository) List(ctx context.Context, offset, limit int, activeOnly bool) ([]Package, int64, error) {
	q := r.db.WithContext(ctx).Model(&Package{})
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var packages []Package
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&packages).Error; err != nil {
		return nil, 0, err
	}
	return packages, total, nil
}

func (r *repository) SlugExists(ctx context.Context, slug, excludeID string) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&Package{}).Where("slug = ?", slug)
	if excludeID != "" {
		q = q.Where("id <> ?", excludeID)
	}
	err := q.Count(&count).Error
	return count > 0, err
}
