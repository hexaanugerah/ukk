package destinationinfo

type CreateRequest struct {
	Title           string `json:"title" validate:"required,min=3,max=200"`
	Description     string `json:"description" validate:"required,min=10"`
	History         string `json:"history" validate:"omitempty"`
	DestinationDesc string `json:"destination_desc" validate:"omitempty"`
	Location        string `json:"location" validate:"omitempty,max=500"`
	OperatingHours  string `json:"operating_hours" validate:"omitempty,max=200"`
	Contact         string `json:"contact" validate:"omitempty,max=500"`
	AdditionalInfo  string `json:"additional_info" validate:"omitempty"`
}

type UpdateRequest struct {
	Title           string `json:"title" validate:"required,min=3,max=200"`
	Description     string `json:"description" validate:"required,min=10"`
	History         string `json:"history" validate:"omitempty"`
	DestinationDesc string `json:"destination_desc" validate:"omitempty"`
	Location        string `json:"location" validate:"omitempty,max=500"`
	OperatingHours  string `json:"operating_hours" validate:"omitempty,max=200"`
	Contact         string `json:"contact" validate:"omitempty,max=500"`
	AdditionalInfo  string `json:"additional_info" validate:"omitempty"`
}

type Response struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	History         string `json:"history"`
	DestinationDesc string `json:"destination_desc"`
	Location        string `json:"location"`
	OperatingHours  string `json:"operating_hours"`
	Contact         string `json:"contact"`
	AdditionalInfo  string `json:"additional_info"`
	Status          string `json:"status"`
	AuthorID        string `json:"author_id"`
	PublishedAt     *int64 `json:"published_at,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func ToResponse(i *DestinationInfo) *Response {
	if i == nil {
		return nil
	}
	return &Response{
		ID:              i.ID,
		Title:           i.Title,
		Description:     i.Description,
		History:         i.History,
		DestinationDesc: i.DestinationDesc,
		Location:        i.Location,
		OperatingHours:  i.OperatingHours,
		Contact:         i.Contact,
		AdditionalInfo:  i.AdditionalInfo,
		Status:          i.Status,
		AuthorID:        i.AuthorID,
		PublishedAt:     i.PublishedAt,
		CreatedAt:       i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       i.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToResponseList(items []*DestinationInfo) []*Response {
	out := make([]*Response, 0, len(items))
	for _, i := range items {
		out = append(out, ToResponse(i))
	}
	return out
}
