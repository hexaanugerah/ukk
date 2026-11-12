package user

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context, offset, limit int, search string) ([]AdminUser, int64, error)
	FindByID(ctx context.Context, id string) (*AdminUser, error)
	SetActive(ctx context.Context, id string, active bool, actorID string) error
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) List(ctx context.Context, offset, limit int, search string) ([]AdminUser, int64, error) {
	q := r.db.WithContext(ctx).Model(&AdminUser{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("name LIKE ? OR email LIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []AdminUser
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*AdminUser, error) {
	var u AdminUser
	err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) SetActive(ctx context.Context, id string, active bool, actorID string) error {
	return r.db.WithContext(ctx).Model(&AdminUser{}).Where("id = ?", id).
		Updates(map[string]interface{}{"is_active": active, "updated_by": actorID}).Error
}
