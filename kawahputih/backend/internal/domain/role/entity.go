package role

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	return nil
}

type Role struct {
	Base
	Name        string       `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description string       `gorm:"type:varchar(255)" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	Base
	Code        string `gorm:"type:varchar(100);uniqueIndex;not null" json:"code"`
	Description string `gorm:"type:varchar(255)" json:"description"`
}

func (Permission) TableName() string { return "permissions" }

// userRole adalah baris tabel join, sengaja tidak diekspor dan tidak dipakai sebagai relasi GORM ke auth.User. Domain role tidak mengimpor struct User, hanya ID-nya. Relasi dimodelkan di level tabel, bukan level tipe Go.
type userRole struct {
	UserID string `gorm:"column:user_id"`
	RoleID string `gorm:"column:role_id"`
}

func (userRole) TableName() string { return "user_roles" }

// Nama role kanonik. String persis inilah yang dibawa claim role JWT dan dicek middleware.RequireRole; cocok dengan union AppRole frontend (config/auth.config.ts) sehingga satu ejaan berlaku end-to-end.
const (
	RoleAdmin      = "ADMIN"
	RoleManager    = "MANAGER"
	RolePetugas    = "PETUGAS"
	RolePengunjung = "PENGUNJUNG"
	// ROLE-20: super_admin adalah satu-satunya role yang boleh mengelola
	// penetapan role via /super-admin/* (lihat middleware.RequireRole di routes).
	RoleSuperAdmin = "super_admin"
)

const DefaultRoleName = RolePengunjung
