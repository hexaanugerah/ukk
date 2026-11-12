package public

import "time"

type DestinationResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Website     string `json:"website"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
	CoverImage  string `json:"cover_image"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func ToDestinationResponse(d *Destination) *DestinationResponse {
	return &DestinationResponse{
		ID:          d.ID,
		Name:        d.Name,
		Slug:        d.Slug,
		Description: d.Description,
		Address:     d.Address,
		Phone:       d.Phone,
		Email:       d.Email,
		Website:     d.Website,
		Latitude:    d.Latitude,
		Longitude:   d.Longitude,
		CoverImage:  d.CoverImage,
		IsActive:    d.IsActive,
		SortOrder:   d.SortOrder,
		CreatedAt:   d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   d.UpdatedAt.Format(time.RFC3339),
	}
}

func ToDestinationResponseList(items []Destination) []DestinationResponse {
	out := make([]DestinationResponse, 0, len(items))
	for i := range items {
		out = append(out, *ToDestinationResponse(&items[i]))
	}
	return out
}

type FacilityResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Category    string `json:"category"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func ToFacilityResponse(f *Facility) *FacilityResponse {
	return &FacilityResponse{
		ID:          f.ID,
		Name:        f.Name,
		Description: f.Description,
		Icon:        f.Icon,
		Category:    f.Category,
		IsActive:    f.IsPublished && f.Status == "active",
		SortOrder:   f.SortOrder,
		CreatedAt:   f.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   f.UpdatedAt.Format(time.RFC3339),
	}
}

func ToFacilityResponseList(items []Facility) []FacilityResponse {
	out := make([]FacilityResponse, 0, len(items))
	for i := range items {
		out = append(out, *ToFacilityResponse(&items[i]))
	}
	return out
}

type FAQResponse struct {
	ID        string `json:"id"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	Category  string `json:"category"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func ToFAQResponse(f *FAQ) *FAQResponse {
	return &FAQResponse{
		ID:        f.ID,
		Question:  f.Question,
		Answer:    f.Answer,
		Category:  f.Category,
		SortOrder: f.SortOrder,
		IsActive:  f.IsPublished,
		CreatedAt: f.CreatedAt.Format(time.RFC3339),
		UpdatedAt: f.UpdatedAt.Format(time.RFC3339),
	}
}

func ToFAQResponseList(items []FAQ) []FAQResponse {
	out := make([]FAQResponse, 0, len(items))
	for i := range items {
		out = append(out, *ToFAQResponse(&items[i]))
	}
	return out
}

type ReviewResponse struct {
	ID            string `json:"id"`
	DestinationID string `json:"destination_id"`
	UserID        string `json:"user_id"`
	Rating        int    `json:"rating"`
	Comment       string `json:"comment"`
	IsApproved    bool   `json:"is_approved"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func ToReviewResponse(r *Review) *ReviewResponse {
	return &ReviewResponse{
		ID:            r.ID,
		DestinationID: r.DestinationID,
		UserID:        r.UserID,
		Rating:        r.Rating,
		Comment:       r.Comment,
		IsApproved:    r.IsApproved,
		CreatedAt:     r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     r.UpdatedAt.Format(time.RFC3339),
	}
}

func ToReviewResponseList(items []Review) []ReviewResponse {
	out := make([]ReviewResponse, 0, len(items))
	for i := range items {
		out = append(out, *ToReviewResponse(&items[i]))
	}
	return out
}

type ContentResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Type        string `json:"type"`
	Excerpt     string `json:"excerpt"`
	Body        string `json:"body"`
	CoverImage  string `json:"cover_image"`
	Status      string `json:"status"`
	AuthorID    string `json:"author_id"`
	PublishedAt *int64 `json:"published_at,omitempty"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func ToContentResponse(c *Content) *ContentResponse {
	return &ContentResponse{
		ID:          c.ID,
		Title:       c.Title,
		Slug:        c.Slug,
		Type:        c.Type,
		Excerpt:     c.Excerpt,
		Body:        c.Body,
		CoverImage:  c.CoverImage,
		Status:      c.Status,
		AuthorID:    c.AuthorID,
		PublishedAt: c.PublishedAt,
		SortOrder:   c.SortOrder,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}
}

func ToContentResponseList(items []Content) []ContentResponse {
	out := make([]ContentResponse, 0, len(items))
	for i := range items {
		out = append(out, *ToContentResponse(&items[i]))
	}
	return out
}

type PriceInfoResponse struct {
	TicketID   string `json:"ticket_id"`
	TicketName string `json:"ticket_name"`
	Category   string `json:"category"`
	Price      int64  `json:"price"`
	Capacity   int    `json:"capacity"`
	Available  int    `json:"available"`
	IsActive   bool   `json:"is_active"`
}
