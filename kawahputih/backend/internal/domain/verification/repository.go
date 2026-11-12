package verification

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, v *Verification) error
	FindByID(ctx context.Context, id string) (*Verification, error)
	List(ctx context.Context, offset, limit int, gate, petugasID, from, to string) ([]Verification, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, v *Verification) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*Verification, error) {
	var v Verification
	err := r.db.WithContext(ctx).First(&v, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *repository) List(ctx context.Context, offset, limit int, gate, petugasID, from, to string) ([]Verification, int64, error) {
	q := r.db.WithContext(ctx).Model(&Verification{}).Order("scanned_at DESC")
	if gate != "" {
		q = q.Where("gate = ?", gate)
	}
	if petugasID != "" {
		q = q.Where("petugas_id = ?", petugasID)
	}
	if from != "" {
		q = q.Where("scanned_at >= ?", from)
	}
	if to != "" {
		q = q.Where("scanned_at <= ?", to)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []Verification
	err := q.Offset(offset).Limit(limit).Find(&items).Error
	return items, total, err
}
