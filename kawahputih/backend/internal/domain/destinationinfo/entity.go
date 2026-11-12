package destinationinfo

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

type DestinationInfo struct {
	Base
	Title           string  `gorm:"type:varchar(200);not null;uniqueIndex" json:"title"`
	Description     string  `gorm:"type:longtext;not null" json:"description"`
	History         string  `gorm:"type:longtext" json:"history"`
	DestinationDesc string  `gorm:"type:longtext" json:"destination_desc"`
	Location        string  `gorm:"type:varchar(500)" json:"location"`
	OperatingHours  string  `gorm:"type:varchar(200)" json:"operating_hours"`
	Contact         string  `gorm:"type:varchar(500)" json:"contact"`
	AdditionalInfo  string  `gorm:"type:longtext" json:"additional_info"`
	Status          string  `gorm:"type:varchar(20);not null;default:draft;index" json:"status"`
	AuthorID        string  `gorm:"type:char(36);index;not null" json:"author_id"`
	CreatedBy       string  `gorm:"type:char(36);index" json:"created_by,omitempty"`
	UpdatedBy       string  `gorm:"type:char(36)" json:"updated_by,omitempty"`
	DeletedBy       *string `gorm:"type:char(36)" json:"-"`
	PublishedAt     *int64  `json:"published_at,omitempty"`
}

func (DestinationInfo) TableName() string { return "destination_information" }

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)
