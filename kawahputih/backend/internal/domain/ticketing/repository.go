package ticketing

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, t *Ticket) error
	Update(ctx context.Context, t *Ticket) error
	// MarkUsedIfValid membakar tiket secara atomik: hanya mengembalikan true
	// bila baris yang berstatus VALID berhasil diubah menjadi USED. Mencegah
	// dua petugas yang memindai QR yang sama secara hampir bersamaan dihitung
	// ganda (double check-in).
	MarkUsedIfValid(ctx context.Context, ticketID, gate string, at time.Time) (bool, error)
	FindByID(ctx context.Context, id string) (*Ticket, error)
	FindByQR(ctx context.Context, qr string) (*Ticket, error)
	FindByBookingID(ctx context.Context, bookingID string) (*Ticket, error)
	ListByUserID(ctx context.Context, userID string) ([]Ticket, error)
	SetStatusForBooking(ctx context.Context, bookingID, status string) (int64, error)
	CountByStatus(ctx context.Context) (map[string]int64, error)
	CountValidOn(ctx context.Context, date string) (int64, error)

	CreateScan(ctx context.Context, s *ScanRecord) error
	ListScans(ctx context.Context, offset, limit int, gate, petugasID, from, to string) ([]ScanRecord, int64, error)
	CountScansByStatus(ctx context.Context) (map[string]int64, error)
	CountScansOn(ctx context.Context, date string) (int64, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, t *Ticket) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *repository) Update(ctx context.Context, t *Ticket) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *repository) MarkUsedIfValid(ctx context.Context, ticketID, gate string, at time.Time) (bool, error) {
	res := r.db.WithContext(ctx).Model(&Ticket{}).
		Where("id = ? AND status = ?", ticketID, TicketValid).
		Updates(map[string]interface{}{
			"status":          TicketUsed,
			"checked_in_at":   at,
			"checked_in_gate": gate,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *repository) findBy(ctx context.Context, where string, arg interface{}) (*Ticket, error) {
	var t Ticket
	err := r.db.WithContext(ctx).First(&t, where, arg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*Ticket, error) {
	return r.findBy(ctx, "id = ?", id)
}

func (r *repository) FindByQR(ctx context.Context, qr string) (*Ticket, error) {
	return r.findBy(ctx, "qr_data = ?", qr)
}

func (r *repository) FindByBookingID(ctx context.Context, bookingID string) (*Ticket, error) {
	return r.findBy(ctx, "booking_id = ?", bookingID)
}

func (r *repository) ListByUserID(ctx context.Context, userID string) ([]Ticket, error) {
	var out []Ticket
	err := r.db.WithContext(ctx).Model(&Ticket{}).
		Joins("JOIN bookings ON bookings.id = tickets.booking_id").
		Where("bookings.user_id = ?", userID).
		Order("tickets.generated_at DESC").
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *repository) SetStatusForBooking(ctx context.Context, bookingID, status string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&Ticket{}).Where("booking_id = ?", bookingID)
	if status == TicketValid {
		// Tiket yang sudah digunakan tidak boleh dihidupkan kembali oleh
		// perubahan status booking (mis. override admin ke PAID/CONFIRMED).
		q = q.Where("status <> ?", TicketUsed)
	}
	res := q.Update("status", status)
	return res.RowsAffected, res.Error
}

func (r *repository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	rows := []struct {
		Status string
		Count  int64
	}{}
	err := r.db.WithContext(ctx).Model(&Ticket{}).
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

func (r *repository) CountValidOn(ctx context.Context, date string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Ticket{}).
		Where("visit_date = ? AND status = ?", date, TicketValid).Count(&count).Error
	return count, err
}

func (r *repository) CreateScan(ctx context.Context, s *ScanRecord) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func scanFilter(db *gorm.DB, gate, petugasID, from, to string) *gorm.DB {
	if gate != "" {
		db = db.Where("gate = ?", gate)
	}
	if petugasID != "" {
		db = db.Where("petugas_id = ?", petugasID)
	}
	if from != "" {
		db = db.Where("scanned_at >= ?", from+" 00:00:00")
	}
	if to != "" {
		db = db.Where("scanned_at <= ?", to+" 23:59:59")
	}
	return db
}

func (r *repository) ListScans(ctx context.Context, offset, limit int, gate, petugasID, from, to string) ([]ScanRecord, int64, error) {
	q := scanFilter(r.db.WithContext(ctx).Model(&ScanRecord{}), gate, petugasID, from, to)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []ScanRecord
	if err := q.Order("scanned_at DESC").Offset(offset).Limit(limit).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *repository) CountScansByStatus(ctx context.Context) (map[string]int64, error) {
	rows := []struct {
		Status string
		Count  int64
	}{}
	err := r.db.WithContext(ctx).Model(&ScanRecord{}).
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

func (r *repository) CountScansOn(ctx context.Context, date string) (int64, error) {
	start, err := time.Parse("2006-01-02", date)
	if err != nil {
		return 0, err
	}
	end := start.Add(24 * time.Hour)
	var count int64
	err = r.db.WithContext(ctx).Model(&ScanRecord{}).
		Where("scanned_at >= ? AND scanned_at < ?", start, end).Count(&count).Error
	return count, err
}

// biarkan time tetap diimpor untuk pemakaian tanggal di interface bila nanti diperlukan
var _ = time.Now
