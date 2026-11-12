package visitorrule

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

type VisitorRule struct {
	Base
	Title       string  `gorm:"type:varchar(200);not null" json:"title"`
	Description string  `gorm:"type:longtext;not null" json:"description"`
	Category    string  `gorm:"type:varchar(100);index;not null" json:"category"`
	Priority    string  `gorm:"type:varchar(20);not null;default:normal;index" json:"priority"`
	SortOrder   int     `gorm:"default:0" json:"sort_order"`
	IsPublished bool    `gorm:"default:false;index" json:"is_published"`
	CreatedBy   string  `gorm:"type:char(36);index" json:"created_by,omitempty"`
	UpdatedBy   string  `gorm:"type:char(36)" json:"updated_by,omitempty"`
	DeletedBy   *string `gorm:"type:char(36)" json:"-"`
}

func (VisitorRule) TableName() string { return "visitor_rules" }

const (
	PriorityLow      = "low"
	PriorityNormal   = "normal"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)
