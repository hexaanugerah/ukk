package tourismpackage

type CreateRequest struct {
	Name          string `json:"name" validate:"required,min=3,max=200"`
	Description   string `json:"description" validate:"max=5000"`
	CoverImage    string `json:"cover_image" validate:"omitempty,url"`
	PriceCents    int64  `json:"price_cents" validate:"required,min=0"`
	DurationHours int    `json:"duration_hours" validate:"required,min=1"`
	MaxCapacity   int    `json:"max_capacity" validate:"min=0"`
}

type UpdateRequest struct {
	Name          string `json:"name" validate:"required,min=3,max=200"`
	Description   string `json:"description" validate:"max=5000"`
	CoverImage    string `json:"cover_image" validate:"omitempty,url"`
	PriceCents    int64  `json:"price_cents" validate:"required,min=0"`
	DurationHours int    `json:"duration_hours" validate:"required,min=1"`
	MaxCapacity   int    `json:"max_capacity" validate:"min=0"`
	IsActive      bool   `json:"is_active"`
}

type Response struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Description   string `json:"description"`
	CoverImage    string `json:"cover_image"`
	PriceCents    int64  `json:"price_cents"`
	Currency      string `json:"currency"`
	DurationHours int    `json:"duration_hours"`
	MaxCapacity   int    `json:"max_capacity"`
	IsActive      bool   `json:"is_active"`
}
