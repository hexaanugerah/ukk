package facility

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

type Facility struct {
	Base
	Name           string  `gorm:"type:varchar(150);not null" json:"name"`
	Description    string  `gorm:"type:varchar(500)" json:"description"`
	Icon           string  `gorm:"type:varchar(200)" json:"icon"`
	ImageURL       string  `gorm:"type:varchar(500)" json:"image_url"`
	Location       string  `gorm:"type:varchar(200)" json:"location"`
	OperatingHours string  `gorm:"type:varchar(200)" json:"operating_hours"`
	Status         string  `gorm:"type:varchar(20);not null;default:active;index" json:"status"`
	IsPublished    bool    `gorm:"default:false;index" json:"is_published"`
	SortOrder      int     `gorm:"default:0" json:"sort_order"`
	CreatedBy      string  `gorm:"type:char(36);index" json:"created_by,omitempty"`
	UpdatedBy      string  `gorm:"type:char(36)" json:"updated_by,omitempty"`
	DeletedBy      *string `gorm:"type:char(36)" json:"-"`
}

func (Facility) TableName() string { return "facilities" }

const (
	StatusActive      = "active"
	StatusMaintenance = "maintenance"
	StatusInactive    = "inactive"
)
