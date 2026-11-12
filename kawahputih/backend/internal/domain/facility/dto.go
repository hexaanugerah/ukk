package facility

type CreateRequest struct {
	Name           string `json:"name" validate:"required,min=3,max=150"`
	Description    string `json:"description" validate:"omitempty,max=500"`
	Icon           string `json:"icon" validate:"omitempty,max=200"`
	ImageURL       string `json:"image_url" validate:"omitempty,url,max=500"`
	Location       string `json:"location" validate:"omitempty,max=200"`
	OperatingHours string `json:"operating_hours" validate:"omitempty,max=200"`
	Status         string `json:"status" validate:"omitempty,oneof=active maintenance inactive"`
	IsPublished    bool   `json:"is_published"`
	SortOrder      int    `json:"sort_order"`
}

type UpdateRequest struct {
	Name           string `json:"name" validate:"required,min=3,max=150"`
	Description    string `json:"description" validate:"omitempty,max=500"`
	Icon           string `json:"icon" validate:"omitempty,max=200"`
	ImageURL       string `json:"image_url" validate:"omitempty,url,max=500"`
	Location       string `json:"location" validate:"omitempty,max=200"`
	OperatingHours string `json:"operating_hours" validate:"omitempty,max=200"`
	Status         string `json:"status" validate:"omitempty,oneof=active maintenance inactive"`
	IsPublished    bool   `json:"is_published"`
	SortOrder      int    `json:"sort_order"`
}

type Response struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Icon           string `json:"icon"`
	ImageURL       string `json:"image_url"`
	Location       string `json:"location"`
	OperatingHours string `json:"operating_hours"`
	Status         string `json:"status"`
	IsPublished    bool   `json:"is_published"`
	SortOrder      int    `json:"sort_order"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func ToResponse(f *Facility) *Response {
	if f == nil {
		return nil
	}
	return &Response{
		ID:             f.ID,
		Name:           f.Name,
		Description:    f.Description,
		Icon:           f.Icon,
		ImageURL:       f.ImageURL,
		Location:       f.Location,
		OperatingHours: f.OperatingHours,
		Status:         f.Status,
		IsPublished:    f.IsPublished,
		SortOrder:      f.SortOrder,
		CreatedAt:      f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToResponseList(items []*Facility) []*Response {
	out := make([]*Response, 0, len(items))
	for _, f := range items {
		out = append(out, ToResponse(f))
	}
	return out
}
