package faq

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

type FAQ struct {
	Base
	Question    string  `gorm:"type:varchar(500);not null" json:"question"`
	Answer      string  `gorm:"type:longtext;not null" json:"answer"`
	Category    string  `gorm:"type:varchar(100);index;not null" json:"category"`
	SortOrder   int     `gorm:"default:0" json:"sort_order"`
	IsPublished bool    `gorm:"default:false;index" json:"is_published"`
	CreatedBy   string  `gorm:"type:char(36);index" json:"created_by,omitempty"`
	UpdatedBy   string  `gorm:"type:char(36)" json:"updated_by,omitempty"`
	DeletedBy   *string `gorm:"type:char(36)" json:"-"`
}

func (FAQ) TableName() string { return "faqs" }
