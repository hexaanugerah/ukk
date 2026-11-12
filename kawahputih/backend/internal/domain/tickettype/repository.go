package tickettype

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, t *TicketType) error
	Update(ctx context.Context, t *TicketType) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*TicketType, error)
	FindByCode(ctx context.Context, code string) (*TicketType, error)
	List(ctx context.Context, offset, limit int, status, category string) ([]TicketType, int64, error)
	ListActive(ctx context.Context) ([]TicketType, error)
	CountCategories(ctx context.Context) ([]string, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, t *TicketType) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *repository) Update(ctx context.Context, t *TicketType) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&TicketType{}, "id = ?", id).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*TicketType, error) {
	var t TicketType
	err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) FindByCode(ctx context.Context, code string) (*TicketType, error) {
	var t TicketType
	err := r.db.WithContext(ctx).First(&t, "code = ?", code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) List(ctx context.Context, offset, limit int, status, category string) ([]TicketType, int64, error) {
	q := r.db.WithContext(ctx).Model(&TicketType{})
	if status == StatusActive {
		q = q.Where("is_active = ?", true)
	} else if status == StatusInactive {
		q = q.Where("is_active = ?", false)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []TicketType
	if err := q.Order("created_at ASC").Offset(offset).Limit(limit).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *repository) ListActive(ctx context.Context) ([]TicketType, error) {
	var out []TicketType
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("created_at ASC").Find(&out).Error
	return out, err
}

func (r *repository) CountCategories(ctx context.Context) ([]string, error) {
	var cats []string
	err := r.db.WithContext(ctx).Model(&TicketType{}).Distinct().Pluck("category", &cats).Error
	return cats, err
}
