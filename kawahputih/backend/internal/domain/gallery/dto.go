package gallery

type CreateRequest struct {
	Title     string `form:"title" validate:"required,min=3,max=200"`
	Category  string `form:"category" validate:"omitempty,max=100"`
	SortOrder int    `form:"sort_order"`
}

type UpdateRequest struct {
	Title     string `form:"title" validate:"required,min=3,max=200"`
	Category  string `form:"category" validate:"omitempty,max=100"`
	SortOrder int    `form:"sort_order"`
}

type Response struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	ImageURL   string `json:"image_url"`
	Category   string `json:"category"`
	SortOrder  int    `json:"sort_order"`
	UploadedBy string `json:"uploaded_by"`
}

func ToResponse(i *Item) *Response {
	if i == nil {
		return nil
	}
	return &Response{
		ID:         i.ID,
		Title:      i.Title,
		ImageURL:   i.ImageURL,
		Category:   i.Category,
		SortOrder:  i.SortOrder,
		UploadedBy: i.UploadedBy,
	}
}

func ToResponseList(items []*Item) []*Response {
	out := make([]*Response, 0, len(items))
	for _, i := range items {
		out = append(out, ToResponse(i))
	}
	return out
}
