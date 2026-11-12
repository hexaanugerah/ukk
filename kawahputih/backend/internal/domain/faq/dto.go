package faq

type CreateRequest struct {
	Question    string `json:"question" validate:"required,min=5,max=500"`
	Answer      string `json:"answer" validate:"required,min=10"`
	Category    string `json:"category" validate:"required,oneof=Tiket Jam Operasional Lokasi Fasilitas Aturan Kunjungan Parkir Pembayaran Umum"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
}

type UpdateRequest struct {
	Question    string `json:"question" validate:"required,min=5,max=500"`
	Answer      string `json:"answer" validate:"required,min=10"`
	Category    string `json:"category" validate:"required,oneof=Tiket Jam Operasional Lokasi Fasilitas Aturan Kunjungan Parkir Pembayaran Umum"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
}

type Response struct {
	ID          string `json:"id"`
	Question    string `json:"question"`
	Answer      string `json:"answer"`
	Category    string `json:"category"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func ToResponse(f *FAQ) *Response {
	if f == nil {
		return nil
	}
	return &Response{
		ID:          f.ID,
		Question:    f.Question,
		Answer:      f.Answer,
		Category:    f.Category,
		SortOrder:   f.SortOrder,
		IsPublished: f.IsPublished,
		CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToResponseList(items []*FAQ) []*Response {
	out := make([]*Response, 0, len(items))
	for _, f := range items {
		out = append(out, ToResponse(f))
	}
	return out
}
