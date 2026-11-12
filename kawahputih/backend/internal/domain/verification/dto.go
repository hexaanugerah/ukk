package verification

import "time"

type VerificationResponse struct {
	ID          string    `json:"id"`
	TicketID    string    `json:"ticket_id"`
	BookingID   string    `json:"booking_id"`
	VisitorName string    `json:"visitor_name"`
	Gate        string    `json:"gate"`
	PetugasID   string    `json:"petugas_id"`
	PetugasName string    `json:"petugas_name"`
	Status      string    `json:"status"`
	ScannedAt   time.Time `json:"scanned_at"`
	Note        string    `json:"note"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

func ToVerificationResponse(v *Verification) *VerificationResponse {
	return &VerificationResponse{
		ID:          v.ID,
		TicketID:    v.TicketID,
		BookingID:   v.BookingID,
		VisitorName: v.VisitorName,
		Gate:        v.Gate,
		PetugasID:   v.PetugasID,
		PetugasName: v.PetugasName,
		Status:      v.Status,
		ScannedAt:   v.ScannedAt,
		Note:        v.Note,
		CreatedAt:   v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   v.UpdatedAt.Format(time.RFC3339),
	}
}

func ToVerificationResponseList(items []Verification) []VerificationResponse {
	out := make([]VerificationResponse, 0, len(items))
	for i := range items {
		out = append(out, *ToVerificationResponse(&items[i]))
	}
	return out
}

type VerifyRequest struct {
	TicketID string `json:"ticket_id" validate:"required,min=4,max=120"`
	Gate     string `json:"gate" validate:"omitempty,max=100"`
	Petugas  string `json:"petugas" validate:"omitempty,max=150"`
	Note     string `json:"note" validate:"omitempty,max=500"`
}

type VerifyResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Status  string          `json:"status"`
	Ticket  *TicketResponse `json:"ticket,omitempty"`
}

type TicketResponse struct {
	ID            string     `json:"id"`
	BookingID     string     `json:"booking_id"`
	VisitorName   string     `json:"visitor_name"`
	Email         string     `json:"email"`
	VisitDate     string     `json:"visit_date"`
	TicketName    string     `json:"ticket_name"`
	Quantity      int        `json:"quantity"`
	Total         int64      `json:"total"`
	Status        string     `json:"status"`
	GeneratedAt   time.Time  `json:"generated_at"`
	CheckedInAt   *time.Time `json:"checked_in_at,omitempty"`
	CheckedInGate string     `json:"checked_in_gate,omitempty"`
	QRData        string     `json:"qr_data"`
}
