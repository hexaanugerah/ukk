package visitorrule

type CreateRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Description string `json:"description" validate:"required,min=10"`
	Category    string `json:"category" validate:"required,oneof=Keselamatan Lingkungan Kebersihan Kendaraan Aktivitas Pengunjung Larangan Informasi Umum"`
	Priority    string `json:"priority" validate:"omitempty,oneof=low normal high critical"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
}

type UpdateRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Description string `json:"description" validate:"required,min=10"`
	Category    string `json:"category" validate:"required,oneof=Keselamatan Lingkungan Kebersihan Kendaraan Aktivitas Pengunjung Larangan Informasi Umum"`
	Priority    string `json:"priority" validate:"omitempty,oneof=low normal high critical"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
}

type Response struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func ToResponse(r *VisitorRule) *Response {
	if r == nil {
		return nil
	}
	return &Response{
		ID:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		Category:    r.Category,
		Priority:    r.Priority,
		SortOrder:   r.SortOrder,
		IsPublished: r.IsPublished,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToResponseList(items []*VisitorRule) []*Response {
	out := make([]*Response, 0, len(items))
	for _, r := range items {
		out = append(out, ToResponse(r))
	}
	return out
}
