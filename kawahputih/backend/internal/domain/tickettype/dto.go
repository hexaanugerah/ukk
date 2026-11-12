package tickettype

type UpsertRequest struct {
	Code        string `json:"code" validate:"omitempty,max=30"`
	Name        string `json:"name" validate:"required,min=3,max=150"`
	Description string `json:"description" validate:"max=500"`
	Category    string `json:"category" validate:"required"`
	Price       int64  `json:"price" validate:"required,min=1"`
	Capacity    int    `json:"capacity" validate:"required,min=1"`
	Status      string `json:"status" validate:"omitempty"`
}

type SetStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// Response cocok satu-satu dengan tipe Ticket frontend (types/api.ts).
type Response struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Price       int64  `json:"price"`
	Status      string `json:"status"`
	Capacity    int    `json:"capacity"`
	Available   int    `json:"available"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
