package visitors

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Repository interface {
	ListVisitors(ctx context.Context) ([]StatsRow, error)
	FindVisitorByID(ctx context.Context, id string) (*StatsRow, error)
	VisitorStats(ctx context.Context) (*VisitorStats, error)

	ListStaff(ctx context.Context) ([]Staff, error)
	FindStaffByID(ctx context.Context, id string) (*Staff, error)
	StaffStats(ctx context.Context) (*StaffStats, error)
	SetStaffActive(ctx context.Context, id string, active bool) error
	CreateStaff(ctx context.Context, in CreateStaffInput) error
}

// Kesalahan spesifik repository agar handler bisa memetakan HTTP status dengan tepat.
var (
	ErrEmailExists  = errors.New("email already registered")
	ErrRoleNotFound = errors.New("role not found")
)

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

const visitorJoin = `JOIN user_roles ON user_roles.user_id = users.id
	JOIN roles ON roles.id = user_roles.role_id`

func (r *repository) visitorRows(ctx context.Context) ([]StatsRow, error) {
	rows := []struct {
		ID        string
		Name      string
		Email     string
		Phone     string
		JoinedAt  time.Time
		Bookings  int64
		Spent     int64
		LastVisit string
	}{}
	err := r.db.WithContext(ctx).
		Table("users").
		Select(`users.id, users.name, users.email, users.phone, users.created_at as joined_at,
			COALESCE(booking_stats.bookings, 0) as bookings,
			COALESCE(booking_stats.spent, 0) as spent,
			booking_stats.last_visit`).
		Joins(visitorJoin).
		Joins(`LEFT JOIN (
			SELECT user_id,
				COUNT(*) as bookings,
				COALESCE(SUM(CASE WHEN status IN ('PAID','CONFIRMED','USED') THEN total ELSE 0 END), 0) as spent,
				MAX(visit_date) as last_visit
			FROM bookings GROUP BY user_id
		) booking_stats ON booking_stats.user_id = users.id`).
		Where("roles.name = ? AND users.deleted_at IS NULL", "PENGUNJUNG").
		Order("users.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]StatsRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, StatsRow{
			Visitor:       Visitor{ID: row.ID, Name: row.Name, Email: row.Email, Phone: row.Phone, JoinedAt: row.JoinedAt},
			TotalBookings: row.Bookings, TotalSpent: row.Spent, LastVisit: row.LastVisit,
		})
	}
	return out, nil
}

func (r *repository) ListVisitors(ctx context.Context) ([]StatsRow, error) {
	return r.visitorRows(ctx)
}

func (r *repository) FindVisitorByID(ctx context.Context, id string) (*StatsRow, error) {
	all, err := r.visitorRows(ctx)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].ID == id {
			return &all[i], nil
		}
	}
	return nil, nil
}

func (r *repository) VisitorStats(ctx context.Context) (*VisitorStats, error) {
	all, err := r.visitorRows(ctx)
	if err != nil {
		return nil, err
	}
	stats := &VisitorStats{TotalVisitors: int64(len(all))}
	firstOfMonth := time.Now().Format("2006-01") + "-01"
	for _, v := range all {
		if v.JoinedAt.Format("2006-01-02") >= firstOfMonth {
			stats.NewThisMonth++
		}
		if v.TotalBookings > 0 {
			stats.ActiveVisitors++
		}
		stats.TotalRevenue += v.TotalSpent
	}
	return stats, nil
}

func (r *repository) staffRows(ctx context.Context) ([]Staff, error) {
	rows := []struct {
		ID          string
		Name        string
		Email       string
		Phone       string
		Gate        string
		Shift       string
		IsActive    bool
		LastLoginAt *time.Time
		Role        string
	}{}
	err := r.db.WithContext(ctx).
		Table("users").
		Select(`users.id, users.name, users.email, users.phone, users.gate, users.shift,
			users.is_active, users.last_login_at as last_login_at, roles.name as role`).
		Joins(visitorJoin).
		Where("roles.name IN ? AND users.deleted_at IS NULL", []string{"ADMIN", "MANAGER", "PETUGAS"}).
		Order("users.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Staff, 0, len(rows))
	for _, row := range rows {
		status := "NONAKTIF"
		if row.IsActive {
			status = "AKTIF"
		}
		out = append(out, Staff{
			ID: row.ID, Name: row.Name, Email: row.Email, Phone: row.Phone,
			Role: row.Role, Gate: row.Gate, Shift: row.Shift,
			IsActive: row.IsActive, Status: status, LastLoginAt: row.LastLoginAt,
		})
	}
	return out, nil
}

func (r *repository) ListStaff(ctx context.Context) ([]Staff, error) { return r.staffRows(ctx) }
func (r *repository) StaffStats(ctx context.Context) (*StaffStats, error) {
	all, err := r.staffRows(ctx)
	if err != nil {
		return nil, err
	}
	stats := &StaffStats{TotalStaff: int64(len(all)), ByRole: make(map[string]int64)}
	for _, st := range all {
		if st.IsActive {
			stats.ActiveStaff++
		}
		stats.ByRole[st.Role]++
		switch st.Role {
		case "ADMIN":
			stats.Admins++
		case "MANAGER":
			stats.Managers++
		case "PETUGAS":
			stats.Petugas++
		}
	}
	return stats, nil
}

func (r *repository) FindStaffByID(ctx context.Context, id string) (*Staff, error) {
	all, err := r.staffRows(ctx)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].ID == id {
			return &all[i], nil
		}
	}
	return nil, nil
}

func (r *repository) SetStaffActive(ctx context.Context, id string, active bool) error {
	res := r.db.WithContext(ctx).Table("users").Where("id = ?", id).Update("is_active", active)
	if res.RowsAffected == 0 {
		return errors.New("staff not found")
	}
	return res.Error
}

// staffUserRow menyisipkan baris baru ke users tanpa mengimpor domain auth,
// mengikuti pola yang sama dengan seed.go.
type staffUserRow struct {
	ID           string `gorm:"type:char(36);primaryKey"`
	Name         string
	Email        string
	Phone        string
	PasswordHash string
	IsActive     bool
	IsVerified   bool
	Gate         string
	Shift        string
}

func (staffUserRow) TableName() string { return "users" }

func (r *repository) CreateStaff(ctx context.Context, in CreateStaffInput) error {
	var count int64
	if err := r.db.WithContext(ctx).Table("users").Where("email = ?", in.Email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrEmailExists
	}

	var roleRow struct {
		ID string
	}
	if err := r.db.WithContext(ctx).Table("roles").Select("id").Where("name = ?", in.Role).Scan(&roleRow).Error; err != nil {
		return err
	}
	if roleRow.ID == "" {
		return ErrRoleNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
	if err != nil {
		return err
	}
	row := staffUserRow{
		ID: uuid.NewString(), Name: in.Name, Email: in.Email, Phone: in.Phone,
		PasswordHash: string(hash), IsActive: true, IsVerified: true,
		Gate: in.Gate, Shift: in.Shift,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", row.ID, roleRow.ID).Error; err != nil {
		return err
	}
	return nil
}
