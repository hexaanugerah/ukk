package auth

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

type User struct {
	Base
	Name         string `gorm:"type:varchar(150);not null" json:"name"`
	Email        string `gorm:"type:varchar(150);uniqueIndex;not null" json:"email"`
	Phone        string `gorm:"type:varchar(20)" json:"phone"`
	PasswordHash string `gorm:"type:varchar(255)" json:"-"`
	// Pointer agar akun non-Google menyimpan NULL; unique index tetap menjaga
	// satu akun per Google ID tanpa bentrok pada banyak akun password.
	GoogleID    *string    `gorm:"type:varchar(100);uniqueIndex" json:"-"`
	AvatarURL   string     `gorm:"type:varchar(500)" json:"avatar_url"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	IsVerified  bool       `gorm:"default:false" json:"is_verified"`
	Gate        string     `gorm:"type:varchar(100)" json:"gate"`
	Shift       string     `gorm:"type:varchar(100)" json:"shift"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

func (User) TableName() string { return "users" }

type RefreshToken struct {
	Base
	UserID    string `gorm:"type:char(36);index;not null" json:"user_id"`
	TokenHash string `gorm:"type:varchar(255);not null" json:"-"`
	ExpiresAt int64  `json:"expires_at"`
	Revoked   bool   `gorm:"default:false" json:"revoked"`
	UserAgent string `gorm:"type:varchar(255)" json:"user_agent"`
	IPAddress string `gorm:"type:varchar(45)" json:"ip_address"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

type AuthToken struct {
	Base
	UserID    string `gorm:"type:char(36);index;not null" json:"user_id"`
	Purpose   string `gorm:"type:varchar(30);not null;index" json:"purpose"`
	TokenHash string `gorm:"type:varchar(255);not null;index" json:"-"`
	ExpiresAt int64  `gorm:"not null" json:"expires_at"`
	UsedAt    *int64 `json:"used_at,omitempty"`
}

func (AuthToken) TableName() string { return "auth_tokens" }

const (
	PurposeVerifyEmail   = "verify_email"
	PurposeResetPassword = "reset_password"
)
