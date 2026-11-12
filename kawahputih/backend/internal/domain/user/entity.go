package user

import "time"

// AdminUser adalah pandangan domain INI atas tabel users, hanya yang dibutuhkan manajemen staf (identitas + status aktif). Tidak memuat PasswordHash, GoogleID, atau hal token. Dua domain memetakan tabel yang sama dengan struct berbeda adalah trade-off yang disengaja.
type AdminUser struct {
	ID         string    `gorm:"type:char(36);primaryKey" json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	IsActive   bool      `json:"is_active"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
}

func (AdminUser) TableName() string { return "users" }
