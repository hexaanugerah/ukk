package payments

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, p *Payment) error
	Update(ctx context.Context, p *Payment) error
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindLatestByBooking(ctx context.Context, bookingID string) (*Payment, error)
	FindByMidtransOrderID(ctx context.Context, orderID string) (*Payment, error)
	FindByEWallet(ctx context.Context, ewallet string) (*Payment, error)
	ListAll(ctx context.Context, filter ListFilter) ([]Payment, int64, error)
	ExpirePendingByBooking(ctx context.Context, bookingID string) (int64, error)
	CountByStatus(ctx context.Context) (map[string]int64, error)
	CountPending(ctx context.Context) (int64, error)
	SumPaidBetween(ctx context.Context, from, to string) (int64, error)
	MethodTotalsBetween(ctx context.Context, from, to string) ([]MethodTotal, error)
	SumSettledBetween(ctx context.Context, from, to string, settled bool) (int64, error)
}

// ListFilter adalah filter daftar pembayaran admin.
type ListFilter struct {
	Offset int
	Limit  int
	Status string
	Method string
	From   string
	To     string
	Query  string
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, p *Payment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *repository) Update(ctx context.Context, p *Payment) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*Payment, error) {
	var p Payment
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repository) FindLatestByBooking(ctx context.Context, bookingID string) (*Payment, error) {
	var p Payment
	err := r.db.WithContext(ctx).Where("booking_id = ?", bookingID).Order("created_at DESC").First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repository) FindByMidtransOrderID(ctx context.Context, orderID string) (*Payment, error) {
	var p Payment
	err := r.db.WithContext(ctx).Where("midtrans_order_id = ?", orderID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repository) FindByEWallet(ctx context.Context, ewallet string) (*Payment, error) {
	var p Payment
	err := r.db.WithContext(ctx).Where("ewallet_number = ?", ewallet).Order("created_at DESC").First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repository) ListAll(ctx context.Context, f ListFilter) ([]Payment, int64, error) {
	q := r.db.WithContext(ctx).Model(&Payment{})
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Method != "" {
		q = q.Where("method = ?", f.Method)
	}
	if f.From != "" {
		q = q.Where("DATE(created_at) >= ?", f.From)
	}
	if f.To != "" {
		q = q.Where("DATE(created_at) <= ?", f.To)
	}
	if f.Query != "" {
		like := "%" + f.Query + "%"
		q = q.Where("id LIKE ? OR booking_id LIKE ? OR reference LIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []Payment
	if err := q.Order("created_at DESC").Offset(f.Offset).Limit(f.Limit).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *repository) ExpirePendingByBooking(ctx context.Context, bookingID string) (int64, error) {
	res := r.db.WithContext(ctx).Model(&Payment{}).
		Where("booking_id = ? AND status IN ?", bookingID, []string{StatusPending, StatusProcessing, StatusFailed}).
		Update("status", StatusExpired)
	return res.RowsAffected, res.Error
}

func (r *repository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	rows := []struct {
		Status string
		Count  int64
	}{}
	err := r.db.WithContext(ctx).Model(&Payment{}).
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

func (r *repository) CountPending(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Payment{}).
		Where("status IN ?", []string{StatusPending, StatusProcessing}).Count(&count).Error
	return count, err
}

func (r *repository) SumPaidBetween(ctx context.Context, from, to string) (int64, error) {
	var sum int64
	err := r.db.WithContext(ctx).Model(&Payment{}).
		Where("status = ? AND DATE(paid_at) >= ? AND DATE(paid_at) <= ?", StatusPaid, from, to).
		Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

func (r *repository) MethodTotalsBetween(ctx context.Context, from, to string) ([]MethodTotal, error) {
	var out []MethodTotal
	err := r.db.WithContext(ctx).Model(&Payment{}).
		Select("method, COUNT(*) as count, COALESCE(SUM(amount), 0) as total").
		Where("status = ? AND DATE(paid_at) >= ? AND DATE(paid_at) <= ?", StatusPaid, from, to).
		Group("method").Order("total DESC").Scan(&out).Error
	return out, err
}

func (r *repository) SumSettledBetween(ctx context.Context, from, to string, settled bool) (int64, error) {
	var sum int64
	status := SettlementUnsettled
	if settled {
		status = SettlementSettled
	}
	err := r.db.WithContext(ctx).Model(&Payment{}).
		Where("status = ? AND settlement_status = ? AND DATE(paid_at) >= ? AND DATE(paid_at) <= ?", StatusPaid, status, from, to).
		Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}
