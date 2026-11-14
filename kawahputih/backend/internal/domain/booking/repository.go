package booking

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	// Create menyimpan booking beserta itemnya dalam satu transaksi.
	Create(ctx context.Context, b *Booking) error
	FindByID(ctx context.Context, id string) (*Booking, error)
	FindByCode(ctx context.Context, code string) (*Booking, error)
	Update(ctx context.Context, b *Booking) error
	ListMine(ctx context.Context, userID string, offset, limit int, status string) ([]Booking, int64, error)
	ListAll(ctx context.Context, offset, limit int, status, search string) ([]Booking, int64, error)
	ListByVisitDate(ctx context.Context, date string, offset, limit int) ([]Booking, int64, error)
	// SoldByTicketOn mengembalikan ticket_id ke total jumlah untuk booking yang memakai kapasitas (PENDING/PAID/CONFIRMED/USED) pada tanggal tertentu. PENDING ikut dihitung agar booking menunggu pembayaran tidak bisa menumpuk melebihi kuota (overbook); booking PENDING kedaluwarsa otomatis sehingga kuota kembali bila pembayaran tidak dilakukan.
	SoldByTicketOn(ctx context.Context, date string) (map[string]int64, error)
	// Agregasi untuk analytics/laporan.
	SumRevenueBetween(ctx context.Context, from, to string) (int64, error)
	CountBetween(ctx context.Context, from, to string, statuses []string) (int64, error)
	CountByStatus(ctx context.Context) (map[string]int64, error)
	StalePending(ctx context.Context, olderThan time.Time, limit int) ([]Booking, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, b *Booking) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Items").Create(b).Error; err != nil {
			return err
		}
		for i := range b.Items {
			b.Items[i].BookingID = b.ID
			if err := tx.Create(&b.Items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repository) preloadItems(db *gorm.DB) *gorm.DB {
	return db.Preload("Items", func(tx *gorm.DB) *gorm.DB { return tx.Order("created_at ASC") })
}

func (r *repository) FindByID(ctx context.Context, id string) (*Booking, error) {
	var b Booking
	err := r.preloadItems(r.db.WithContext(ctx)).First(&b, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) FindByCode(ctx context.Context, code string) (*Booking, error) {
	var b Booking
	err := r.preloadItems(r.db.WithContext(ctx)).Where("booking_code = ?", code).First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) Update(ctx context.Context, b *Booking) error {
	return r.db.WithContext(ctx).Save(b).Error
}

func statusFilter(db *gorm.DB, status string) *gorm.DB {
	if status != "" {
		return db.Where("status = ?", status)
	}
	return db
}

func (r *repository) ListMine(ctx context.Context, userID string, offset, limit int, status string) ([]Booking, int64, error) {
	q := statusFilter(r.db.WithContext(ctx).Model(&Booking{}).Where("user_id = ?", userID), status)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []Booking
	if err := r.preloadItems(q.Order("created_at DESC")).Offset(offset).Limit(limit).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *repository) ListAll(ctx context.Context, offset, limit int, status, search string) ([]Booking, int64, error) {
	q := statusFilter(r.db.WithContext(ctx).Model(&Booking{}), status)
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("booking_code LIKE ? OR visitor_name LIKE ? OR email LIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []Booking
	if err := r.preloadItems(q.Order("created_at DESC")).Offset(offset).Limit(limit).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *repository) ListByVisitDate(ctx context.Context, date string, offset, limit int) ([]Booking, int64, error) {
	q := r.db.WithContext(ctx).Model(&Booking{}).Where("visit_date = ?", date)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []Booking
	if err := r.preloadItems(q.Order("created_at DESC")).Offset(offset).Limit(limit).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *repository) SoldByTicketOn(ctx context.Context, date string) (map[string]int64, error) {
	rows := []struct {
		TicketID string
		Quantity int64
	}{}
	err := r.db.WithContext(ctx).
		Table("booking_items").
		Select("booking_items.ticket_id as ticket_id, SUM(booking_items.quantity) as quantity").
		Joins("JOIN bookings ON bookings.id = booking_items.booking_id").
		Where("bookings.visit_date = ? AND bookings.status IN ?", date, QuotaStatuses).
		Group("booking_items.ticket_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.TicketID] = row.Quantity
	}
	return out, nil
}

func (r *repository) SumRevenueBetween(ctx context.Context, from, to string) (int64, error) {
	var sum int64
	err := r.db.WithContext(ctx).Model(&Booking{}).
		Where("visit_date >= ? AND visit_date <= ? AND status IN ?", from, to, RevenueStatuses).
		Select("COALESCE(SUM(total), 0)").
		Scan(&sum).Error
	return sum, err
}

func (r *repository) CountBetween(ctx context.Context, from, to string, statuses []string) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&Booking{}).Where("visit_date >= ? AND visit_date <= ?", from, to)
	if len(statuses) > 0 {
		q = q.Where("status IN ?", statuses)
	}
	err := q.Count(&count).Error
	return count, err
}

func (r *repository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	rows := []struct {
		Status string
		Count  int64
	}{}
	err := r.db.WithContext(ctx).Model(&Booking{}).
		Select("status, COUNT(*) as count").Group("status").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.Status] = row.Count
	}
	return out, nil
}

func (r *repository) StalePending(ctx context.Context, olderThan time.Time, limit int) ([]Booking, error) {
	var out []Booking
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at <= ?", StatusPending, olderThan).
		Order("created_at ASC").Limit(limit).Find(&out).Error
	return out, err
}
