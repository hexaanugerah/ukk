package tourismpackage

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

// Package bernama tourismpackage di level Go (kata package adalah kata kunci) tapi foldernya tetap internal/domain/package.
type Package struct {
	Base
	Name          string  `gorm:"type:varchar(200);not null" json:"name"`
	Slug          string  `gorm:"type:varchar(220);uniqueIndex;not null" json:"slug"`
	Description   string  `gorm:"type:text" json:"description"`
	CoverImage    string  `gorm:"type:varchar(500)" json:"cover_image"`
	PriceCents    int64   `gorm:"not null" json:"price_cents"`
	Currency      string  `gorm:"type:varchar(3);not null;default:IDR" json:"currency"`
	DurationHours int     `gorm:"not null;default:1" json:"duration_hours"`
	MaxCapacity   int     `gorm:"not null;default:0" json:"max_capacity"`
	IsActive      bool    `gorm:"not null;default:true;index" json:"is_active"`
	CreatedBy     string  `gorm:"type:char(36);index" json:"created_by,omitempty"`
	UpdatedBy     string  `gorm:"type:char(36)" json:"updated_by,omitempty"`
	DeletedBy     *string `gorm:"type:char(36)" json:"-"`
}

func (Package) TableName() string { return "tourism_packages" }
