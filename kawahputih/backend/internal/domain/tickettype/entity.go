package tickettype

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

type TicketType struct {
	Base
	Code        string `gorm:"type:varchar(30);uniqueIndex;not null" json:"code"`
	Name        string `gorm:"type:varchar(150);not null" json:"name"`
	Description string `gorm:"type:varchar(500)" json:"description"`
	Category    string `gorm:"type:varchar(30);not null;default:ENTRANCE;index" json:"category"`
	Price       int64  `gorm:"not null" json:"price"` // IDR
	Capacity    int    `gorm:"not null;default:0" json:"capacity"`
	IsActive    bool   `gorm:"not null;index" json:"is_active"`
	CreatedBy   string `gorm:"type:char(36);index" json:"created_by,omitempty"`
	UpdatedBy   string `gorm:"type:char(36)" json:"updated_by,omitempty"`
}

func (TicketType) TableName() string { return "ticket_types" }

const (
	StatusActive   = "AKTIF"
	StatusInactive = "NONAKTIF"
)

var Categories = []string{"ENTRANCE", "SUNRISE", "TRANSPORT", "PACKAGE", "ACTIVITY", "SERVICE"}
