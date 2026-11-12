package booking

import "time"

// CreateRequest adalah CreateBookingRequest frontend (types/api.ts). Harga tidak pernah diterima dari klien karena backend menghitung ulang tiap baris dari tabel ticket_types.
type CreateRequest struct {
	VisitDate string `json:"visit_date" validate:"required,len=10"`
	Visitor   struct {
		Name  string `json:"name" validate:"required,min=2,max=150"`
		Email string `json:"email" validate:"required,email,max=150"`
		Phone string `json:"phone" validate:"required,min=8,max=20"`
	} `json:"visitor" validate:"required"`
	Items []CreateItem `json:"items" validate:"required,min=1,max=20"`
	Note  string       `json:"note" validate:"omitempty,max=500"`
}

type CreateItem struct {
	TicketID string `json:"ticket_id" validate:"required,len=36"`
	Quantity int    `json:"quantity" validate:"required,min=1,max=50"`
}

// BookingResponse cocok satu-satu dengan tipe Booking frontend.
type BookingResponse struct {
	ID            string         `json:"id"`
	UserID        string         `json:"user_id,omitempty"`
	BookingCode   string         `json:"booking_code,omitempty"`
	VisitorName   string         `json:"visitor_name"`
	Email         string         `json:"email"`
	Phone         string         `json:"phone"`
	VisitDate     string         `json:"visit_date"`
	Items         []ItemResponse `json:"items"`
	Subtotal      int64          `json:"subtotal"`
	Discount      int64          `json:"discount"`
	ServiceFee    int64          `json:"service_fee"`
	Tax           int64          `json:"tax"`
	Total         int64          `json:"total"`
	TotalTickets  int            `json:"total_tickets"`
	Status        string         `json:"status"`
	Note          string         `json:"note,omitempty"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
	CheckedInAt   string         `json:"checked_in_at,omitempty"`
	CheckedInGate string         `json:"checked_in_gate,omitempty"`
	Payment       *PaymentView   `json:"payment,omitempty"`
}

type ItemResponse struct {
	TicketID   string `json:"ticket_id"`
	TicketName string `json:"ticket_name"`
	Price      int64  `json:"price"`
	Quantity   int    `json:"quantity"`
}

// PaymentView adalah snapshot pembayaran minimal yang disematkan di respons booking. Diisi domain payments lewat PaymentStarter.
type PaymentView struct {
	ID        string    `json:"id"`
	BookingID string    `json:"booking_id"`
	Method    string    `json:"method"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// UpdateStatusRequest adalah payload patch status oleh admin.
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// CreateBookingResult adalah payload endpoint create: booking baru plus pembayaran pending yang siap dibayar.
type CreateBookingResult struct {
	Booking BookingResponse `json:"booking"`
	Payment *PaymentView    `json:"payment"`
}
