package gallery

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

type Item struct {
	Base
	Title      string  `gorm:"type:varchar(200);not null" json:"title"`
	ImageURL   string  `gorm:"type:varchar(500);not null" json:"image_url"`
	Category   string  `gorm:"type:varchar(100);index" json:"category"`
	SortOrder  int     `gorm:"default:0" json:"sort_order"`
	UploadedBy string  `gorm:"type:char(36);index;not null" json:"uploaded_by"` // this table's created_by, under a more descriptive name
	UpdatedBy  string  `gorm:"type:char(36)" json:"updated_by,omitempty"`
	DeletedBy  *string `gorm:"type:char(36)" json:"-"`
}

func (Item) TableName() string { return "gallery_items" }
