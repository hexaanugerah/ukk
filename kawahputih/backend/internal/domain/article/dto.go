package article

type CreateRequest struct {
	Title      string `json:"title" validate:"required,min=5,max=200"`
	Excerpt    string `json:"excerpt" validate:"max=500"`
	Content    string `json:"content" validate:"required,min=20"`
	CoverImage string `json:"cover_image" validate:"omitempty,url"`
}

type UpdateRequest struct {
	Title      string `json:"title" validate:"required,min=5,max=200"`
	Excerpt    string `json:"excerpt" validate:"max=500"`
	Content    string `json:"content" validate:"required,min=20"`
	CoverImage string `json:"cover_image" validate:"omitempty,url"`
}

type Response struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Excerpt     string `json:"excerpt"`
	Content     string `json:"content"`
	CoverImage  string `json:"cover_image"`
	Status      string `json:"status"`
	AuthorID    string `json:"author_id"`
	PublishedAt *int64 `json:"published_at,omitempty"`
}
