package article

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

type Article struct {
	Base
	Title       string  `gorm:"type:varchar(200);not null" json:"title"`
	Slug        string  `gorm:"type:varchar(220);uniqueIndex;not null" json:"slug"`
	Excerpt     string  `gorm:"type:varchar(500)" json:"excerpt"`
	Content     string  `gorm:"type:longtext;not null" json:"content"`
	CoverImage  string  `gorm:"type:varchar(500)" json:"cover_image"`
	Status      string  `gorm:"type:varchar(20);not null;default:draft;index" json:"status"`
	AuthorID    string  `gorm:"type:char(36);index;not null" json:"author_id"`
	CreatedBy   string  `gorm:"type:char(36);index" json:"created_by,omitempty"`
	UpdatedBy   string  `gorm:"type:char(36)" json:"updated_by,omitempty"`
	DeletedBy   *string `gorm:"type:char(36)" json:"-"`
	PublishedAt *int64  `json:"published_at,omitempty"`
}

func (Article) TableName() string { return "articles" }

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)
