package visitors

import "time"

// Visitor adalah pandangan admin atas pengguna PENGUNJUNG plus agregat komersialnya. Dipetakan ke users; domain ini tidak mengimpor domain auth.
type Visitor struct {
	ID       string    `gorm:"type:char(36);primaryKey" json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Phone    string    `json:"phone"`
	JoinedAt time.Time `gorm:"->" json:"joined_at"`
	IsActive bool      `json:"-"`
}

func (Visitor) TableName() string { return "users" }

// StatsRow adalah satu pengunjung yang diperkaya dengan agregat booking.
type StatsRow struct {
	Visitor
	TotalBookings int64  `json:"total_bookings"`
	TotalSpent    int64  `json:"total_spent"`
	LastVisit     string `json:"last_visit,omitempty"`
}

// VisitorStats adalah ringkasan halaman pengunjung admin.
type VisitorStats struct {
	TotalVisitors  int64 `json:"total_visitors"`
	NewThisMonth   int64 `json:"new_this_month"`
	ActiveVisitors int64 `json:"active_visitors"` // has at least one PAID booking
	TotalRevenue   int64 `json:"total_revenue"`
}

// Staff adalah pandangan admin atas akun staf (ADMIN/MANAGER/PETUGAS).
type Staff struct {
	ID          string     `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	Role        string     `gorm:"-" json:"role"`
	Gate        string     `json:"gate"`
	Shift       string     `json:"shift"`
	IsActive    bool       `json:"is_active"`
	Status      string     `gorm:"-" json:"status"` // AKTIF | NONAKTIF (derived)
	LastLoginAt *time.Time `json:"last_login"`
}

func (Staff) TableName() string { return "users" }

// StaffStats adalah ringkasan halaman staf admin.
type StaffStats struct {
	TotalStaff  int64            `json:"total_staff"`
	ActiveStaff int64            `json:"active_staff"`
	ByRole      map[string]int64 `json:"by_role"`
	Admins      int64            `json:"admins"`
	Managers    int64            `json:"managers"`
	Petugas     int64            `json:"petugas"`
}
