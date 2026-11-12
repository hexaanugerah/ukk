package verification

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

type Verification struct {
	Base
	TicketID    string    `gorm:"type:char(36);index;not null" json:"ticket_id"`
	BookingID   string    `gorm:"type:char(36);index;not null" json:"booking_id"`
	VisitorName string    `gorm:"type:varchar(150);not null" json:"visitor_name"`
	Gate        string    `gorm:"type:varchar(100)" json:"gate"`
	PetugasID   string    `gorm:"type:char(36);index" json:"petugas_id"`
	PetugasName string    `gorm:"type:varchar(150)" json:"petugas_name"`
	Status      string    `gorm:"type:varchar(20);index;not null" json:"status"`
	ScannedAt   time.Time `gorm:"index;not null" json:"scanned_at"`
	Note        string    `gorm:"type:varchar(500)" json:"note"`
}

func (Verification) TableName() string { return "verifications" }

const (
	StatusValid     = "VALID"
	StatusUsed      = "USED"
	StatusInvalid   = "INVALID"
	StatusExpired   = "EXPIRED"
	StatusCancelled = "CANCELLED"
)
