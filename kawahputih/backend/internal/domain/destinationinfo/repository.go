package destinationinfo

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, info *DestinationInfo) error
	Update(ctx context.Context, info *DestinationInfo) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*DestinationInfo, error)
	FindByTitle(ctx context.Context, title string) (*DestinationInfo, error)
	List(ctx context.Context, offset, limit int, status string) ([]*DestinationInfo, int64, error)
	GetPublished(ctx context.Context) (*DestinationInfo, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, info *DestinationInfo) error {
	return r.db.WithContext(ctx).Create(info).Error
}

func (r *repository) Update(ctx context.Context, info *DestinationInfo) error {
	return r.db.WithContext(ctx).Save(info).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&DestinationInfo{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*DestinationInfo, error) {
	var info DestinationInfo
	err := r.db.WithContext(ctx).First(&info, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (r *repository) FindByTitle(ctx context.Context, title string) (*DestinationInfo, error) {
	var info DestinationInfo
	err := r.db.WithContext(ctx).Where("title = ?", title).First(&info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (r *repository) List(ctx context.Context, offset, limit int, status string) ([]*DestinationInfo, int64, error) {
	q := r.db.WithContext(ctx).Model(&DestinationInfo{}).Order("created_at DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []*DestinationInfo
	if err := q.Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repository) GetPublished(ctx context.Context) (*DestinationInfo, error) {
	var info DestinationInfo
	err := r.db.WithContext(ctx).Where("status = ?", StatusPublished).Order("published_at DESC").First(&info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}
