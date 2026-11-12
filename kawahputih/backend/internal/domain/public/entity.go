package public

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

type Destination struct {
	Base
	Name        string `gorm:"type:varchar(200);not null" json:"name"`
	Slug        string `gorm:"type:varchar(220);uniqueIndex;not null" json:"slug"`
	Description string `gorm:"type:text" json:"description"`
	Address     string `gorm:"type:varchar(500)" json:"address"`
	Phone       string `gorm:"type:varchar(50)" json:"phone"`
	Email       string `gorm:"type:varchar(150)" json:"email"`
	Website     string `gorm:"type:varchar(200)" json:"website"`
	Latitude    string `gorm:"type:varchar(50)" json:"latitude"`
	Longitude   string `gorm:"type:varchar(50)" json:"longitude"`
	CoverImage  string `gorm:"type:varchar(500)" json:"cover_image"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
}

func (Destination) TableName() string { return "destinations" }

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
	Category       string  `gorm:"type:varchar(50);index" json:"category"`
	SortOrder      int     `gorm:"default:0" json:"sort_order"`
	CreatedBy      string  `gorm:"type:char(36);index" json:"created_by,omitempty"`
	UpdatedBy      string  `gorm:"type:char(36)" json:"updated_by,omitempty"`
	DeletedBy      *string `gorm:"type:char(36)" json:"-"`
}

func (Facility) TableName() string { return "facilities" }

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

type Review struct {
	Base
	DestinationID string `gorm:"type:char(36);index" json:"destination_id"`
	UserID        string `gorm:"type:char(36);index" json:"user_id"`
	Rating        int    `gorm:"not null" json:"rating"`
	Comment       string `gorm:"type:text" json:"comment"`
	IsApproved    bool   `gorm:"default:false;index" json:"is_approved"`
}

func (Review) TableName() string { return "reviews" }

type Content struct {
	Base
	Title       string `gorm:"type:varchar(200);not null" json:"title"`
	Slug        string `gorm:"type:varchar(220);uniqueIndex;not null" json:"slug"`
	Type        string `gorm:"type:varchar(50);index;not null" json:"type"`
	Excerpt     string `gorm:"type:varchar(500)" json:"excerpt"`
	Body        string `gorm:"type:longtext" json:"body"`
	CoverImage  string `gorm:"type:varchar(500)" json:"cover_image"`
	Status      string `gorm:"type:varchar(20);index;default:draft" json:"status"`
	AuthorID    string `gorm:"type:char(36);index" json:"author_id"`
	PublishedAt *int64 `json:"published_at,omitempty"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
}

func (Content) TableName() string { return "contents" }

const (
	ContentTypeArticle   = "article"
	ContentTypeEvent     = "event"
	ContentTypePromotion = "promotion"
	ContentTypePage      = "page"
)

const (
	ContentStatusDraft     = "draft"
	ContentStatusPublished = "published"
	ContentStatusArchived  = "archived"
)
