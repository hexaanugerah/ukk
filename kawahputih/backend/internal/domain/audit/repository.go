package audit

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, offset, limit int, actorID, action, targetType, targetID, status, from, to string) ([]AuditLog, int64, error)
	GetByID(ctx context.Context, id string) (*AuditLog, error)
	CountByAction(ctx context.Context, from, to string) (map[string]int64, error)
	CountByActor(ctx context.Context, from, to string) (map[string]int64, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, log *AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *repository) List(ctx context.Context, offset, limit int, actorID, action, targetType, targetID, status, from, to string) ([]AuditLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&AuditLog{})

	if actorID != "" {
		q = q.Where("actor_id = ?", actorID)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if targetType != "" {
		q = q.Where("target_type = ?", targetType)
	}
	if targetID != "" {
		q = q.Where("target_id = ?", targetID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if from != "" {
		q = q.Where("created_at >= ?", from+" 00:00:00")
	}
	if to != "" {
		q = q.Where("created_at <= ?", to+" 23:59:59")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var out []AuditLog
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*AuditLog, error) {
	var log AuditLog
	err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &log, err
}

func (r *repository) CountByAction(ctx context.Context, from, to string) (map[string]int64, error) {
	q := r.db.WithContext(ctx).Model(&AuditLog{})
	if from != "" {
		q = q.Where("created_at >= ?", from+" 00:00:00")
	}
	if to != "" {
		q = q.Where("created_at <= ?", to+" 23:59:59")
	}

	rows := []struct {
		Action string
		Count  int64
	}{}
	err := q.Select("action, COUNT(*) as count").Group("action").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.Action] = row.Count
	}
	return out, nil
}

func (r *repository) CountByActor(ctx context.Context, from, to string) (map[string]int64, error) {
	q := r.db.WithContext(ctx).Model(&AuditLog{})
	if from != "" {
		q = q.Where("created_at >= ?", from+" 00:00:00")
	}
	if to != "" {
		q = q.Where("created_at <= ?", to+" 23:59:59")
	}

	rows := []struct {
		ActorID string
		Count   int64
	}{}
	err := q.Select("actor_id, COUNT(*) as count").Group("actor_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.ActorID] = row.Count
	}
	return out, nil
}
