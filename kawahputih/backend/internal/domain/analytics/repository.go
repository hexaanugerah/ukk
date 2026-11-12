package analytics

import (
	"context"

	"gorm.io/gorm"
)

// dailyRow adalah agregasi mentah per tanggal kunjungan atas booking.
type dailyRow struct {
	VisitDate string
	Bookings  int64
	Tickets   int64
	Revenue   int64
}

// itemRow adalah satu titik performa per jenis tiket.
type itemRow struct {
	Name  string
	Quant int64
}

// statusRow adalah pasangan status/jumlah yang umum.
type statusRow struct {
	Status string
	Count  int64
}

type Repository interface {
	BookingsByDateRange(ctx context.Context, from, to string) ([]dailyRow, error)
	TicketPerformance(ctx context.Context, from, to string, limit int) ([]itemRow, error)
	PaymentsByStatus(ctx context.Context, from, to string) (map[string]int64, error)
	TotalBookings(ctx context.Context) (int64, error)
	BookingsCreatedOn(ctx context.Context, date string) (int64, error)
	TotalTicketsSold(ctx context.Context) (int64, error)
	TotalRevenue(ctx context.Context) (int64, error)
	TotalRevenueBetweenOn(ctx context.Context, from, to string) (int64, error)
	TicketsAvailable(ctx context.Context) (int64, error)
	PendingPayments(ctx context.Context) (int64, error)
	VisitorsCount(ctx context.Context) (int64, error)
	VisitorsInRange(ctx context.Context, from, to string) (int64, error)
	VisitorsOnVisitDate(ctx context.Context, date string) (int64, error)
	ActiveStaffCount(ctx context.Context) (int64, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

// revenueStatuses meniru booking.RevenueStatuses tanpa mengimpor package booking (analytics memakai SQL-nya sendiri).
var revenueStatuses = []string{"PAID", "CONFIRMED", "USED"}

func (r *repository) BookingsByDateRange(ctx context.Context, from, to string) ([]dailyRow, error) {
	rows := []dailyRow{}
	err := r.db.WithContext(ctx).Table("bookings").
		Select("visit_date, COUNT(*) as bookings, COALESCE(SUM(total_tickets),0) as tickets, COALESCE(SUM(total),0) as revenue").
		Where("visit_date >= ? AND visit_date <= ? AND status IN ?", from, to, revenueStatuses).
		Group("visit_date").Order("visit_date ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *repository) TicketPerformance(ctx context.Context, from, to string, limit int) ([]itemRow, error) {
	rows := []itemRow{}
	err := r.db.WithContext(ctx).Table("booking_items").
		Select("booking_items.ticket_name as name, SUM(booking_items.quantity) as quant").
		Joins("JOIN bookings ON bookings.id = booking_items.booking_id").
		Where("bookings.visit_date >= ? AND bookings.visit_date <= ? AND bookings.status IN ?", from, to, revenueStatuses).
		Group("booking_items.ticket_name").
		Order("quant DESC").Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *repository) PaymentsByStatus(ctx context.Context, from, to string) (map[string]int64, error) {
	rows := []statusRow{}
	err := r.db.WithContext(ctx).Table("payments").
		Select("status, COUNT(*) as count").
		Where("DATE(created_at) >= ? AND DATE(created_at) <= ?", from, to).
		Group("status").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.Status] = row.Count
	}
	return out, nil
}

func (r *repository) TotalBookings(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("bookings").Count(&count).Error
	return count, err
}

func (r *repository) BookingsCreatedOn(ctx context.Context, date string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("bookings").
		Where("DATE(created_at) = ?", date).Count(&count).Error
	return count, err
}

func (r *repository) TotalTicketsSold(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("booking_items").
		Select("COALESCE(SUM(booking_items.quantity),0)").
		Joins("JOIN bookings ON bookings.id = booking_items.booking_id").
		Where("bookings.status IN ?", revenueStatuses).
		Scan(&total).Error
	return total, err
}

func (r *repository) TotalRevenue(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("bookings").
		Select("COALESCE(SUM(total),0)").
		Where("status IN ?", revenueStatuses).Scan(&total).Error
	return total, err
}

func (r *repository) TotalRevenueBetweenOn(ctx context.Context, from, to string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("bookings").
		Select("COALESCE(SUM(total),0)").
		Where("visit_date >= ? AND visit_date <= ? AND status IN ?", from, to, revenueStatuses).
		Scan(&total).Error
	return total, err
}

func (r *repository) TicketsAvailable(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("ticket_types").
		Select("COALESCE(SUM(capacity),0)").
		Where("is_active = ?", true).Scan(&total).Error
	return total, err
}

func (r *repository) PendingPayments(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("payments").
		Where("status IN ?", []string{"PENDING", "PROCESSING"}).Count(&count).Error
	return count, err
}

func (r *repository) VisitorsCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("users").
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name = ? AND users.deleted_at IS NULL", "PENGUNJUNG").
		Count(&count).Error
	return count, err
}

// VisitorsInRange menghitung pengunjung unik (user_id berbeda) yang memiliki booking berstatus revenue pada rentang kunjungan.
func (r *repository) VisitorsInRange(ctx context.Context, from, to string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("bookings").
		Select("COUNT(DISTINCT user_id)").
		Where("visit_date >= ? AND visit_date <= ? AND status IN ? AND user_id IS NOT NULL AND user_id <> ''", from, to, revenueStatuses).
		Scan(&count).Error
	return count, err
}

// VisitorsOnVisitDate menjumlahkan tiket terjual untuk kunjungan pada tanggal tertentu (proksi pengunjung hari itu).
func (r *repository) VisitorsOnVisitDate(ctx context.Context, date string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("bookings").
		Select("COALESCE(SUM(total_tickets),0)").
		Where("visit_date = ? AND status IN ?", date, revenueStatuses).
		Scan(&total).Error
	return total, err
}

func (r *repository) ActiveStaffCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("users").
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name IN ? AND users.is_active = ? AND users.deleted_at IS NULL",
			[]string{"ADMIN", "MANAGER", "PETUGAS"}, true).
		Count(&count).Error
	return count, err
}
