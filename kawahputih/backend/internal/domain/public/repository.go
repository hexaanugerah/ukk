package public

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	// Destinasi
	ListDestinations(ctx context.Context, activeOnly bool) ([]Destination, error)
	FindDestinationByID(ctx context.Context, id string) (*Destination, error)
	FindDestinationBySlug(ctx context.Context, slug string) (*Destination, error)

	// Fasilitas
	ListFacilities(ctx context.Context, category string, activeOnly bool) ([]Facility, error)

	// FAQ
	ListFAQs(ctx context.Context, category string, activeOnly bool) ([]FAQ, error)

	// Ulasan
	ListReviews(ctx context.Context, destinationID string, approvedOnly bool, limit int) ([]Review, error)

	// Konten
	ListContents(ctx context.Context, contentType, status string, limit, offset int) ([]Content, int64, error)
	FindContentBySlug(ctx context.Context, slug string) (*Content, error)
	FindContentByID(ctx context.Context, id string) (*Content, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListDestinations(ctx context.Context, activeOnly bool) ([]Destination, error) {
	q := r.db.WithContext(ctx).Model(&Destination{}).Order("sort_order ASC, created_at DESC")
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var items []Destination
	err := q.Find(&items).Error
	return items, err
}

func (r *repository) FindDestinationByID(ctx context.Context, id string) (*Destination, error) {
	var d Destination
	err := r.db.WithContext(ctx).First(&d, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &d, err
}

func (r *repository) FindDestinationBySlug(ctx context.Context, slug string) (*Destination, error) {
	var d Destination
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &d, err
}

func (r *repository) ListFacilities(ctx context.Context, category string, activeOnly bool) ([]Facility, error) {
	q := r.db.WithContext(ctx).Model(&Facility{}).Order("category ASC, sort_order ASC, created_at DESC")
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var items []Facility
	err := q.Find(&items).Error
	return items, err
}

func (r *repository) ListFAQs(ctx context.Context, category string, activeOnly bool) ([]FAQ, error) {
	q := r.db.WithContext(ctx).Model(&FAQ{}).Order("category ASC, sort_order ASC, created_at DESC")
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var items []FAQ
	err := q.Find(&items).Error
	return items, err
}

func (r *repository) ListReviews(ctx context.Context, destinationID string, approvedOnly bool, limit int) ([]Review, error) {
	q := r.db.WithContext(ctx).Model(&Review{}).Order("created_at DESC")
	if destinationID != "" {
		q = q.Where("destination_id = ?", destinationID)
	}
	if approvedOnly {
		q = q.Where("is_approved = ?", true)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	var items []Review
	err := q.Find(&items).Error
	return items, err
}

func (r *repository) ListContents(ctx context.Context, contentType, status string, limit, offset int) ([]Content, int64, error) {
	q := r.db.WithContext(ctx).Model(&Content{}).Order("sort_order ASC, created_at DESC")
	if contentType != "" {
		q = q.Where("type = ?", contentType)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	} else {
		q = q.Where("status = ?", ContentStatusPublished)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []Content
	err := q.Offset(offset).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *repository) FindContentBySlug(ctx context.Context, slug string) (*Content, error) {
	var c Content
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *repository) FindContentByID(ctx context.Context, id string) (*Content, error) {
	var c Content
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}
