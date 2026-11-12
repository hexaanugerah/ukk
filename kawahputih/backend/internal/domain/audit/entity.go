package audit

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	return nil
}

// AuditLog mencatat setiap operasi bisnis penting untuk kepatuhan, debugging, dan keamanan. Tidak bisa diubah setelah ditulis.
type AuditLog struct {
	Base
	ActorID    string `gorm:"type:char(36);index;not null" json:"actor_id"`       // User who performed the action
	ActorRole  string `gorm:"type:varchar(30);not null" json:"actor_role"`        // Role at time of action
	Action     string `gorm:"type:varchar(50);index;not null" json:"action"`      // e.g. "booking.create", "payment.confirm", "ticket.validate"
	TargetType string `gorm:"type:varchar(30);index;not null" json:"target_type"` // e.g. "booking", "payment", "ticket", "ticket_type"
	TargetID   string `gorm:"type:char(36);index;not null" json:"target_id"`      // ID of the affected entity
	Status     string `gorm:"type:varchar(20);not null" json:"status"`            // SUCCESS, FAILED, ERROR
	Metadata   string `gorm:"type:json" json:"metadata,omitempty"`                // JSON blob with context (amounts, old/new values, etc.)
	IPAddress  string `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent  string `gorm:"type:varchar(500)" json:"user_agent,omitempty"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// Konstanta aksi agar konsisten antar domain
const (
	// Aksi booking
	ActionBookingCreate = "booking.create"
	ActionBookingGet    = "booking.get"
	ActionBookingList   = "booking.list"
	ActionBookingCancel = "booking.cancel"
	ActionBookingStatus = "booking.status_change"

	// Aksi pembayaran
	ActionPaymentStart    = "payment.start"
	ActionPaymentProcess  = "payment.process"
	ActionPaymentExpire   = "payment.expire"
	ActionPaymentStatus   = "payment.status_change"
	ActionPaymentDownload = "payment.download"

	// Aksi tiket/validasi
	ActionTicketIssue    = "ticket.issue"
	ActionTicketValidate = "ticket.validate"
	ActionTicketLookup   = "ticket.lookup"
	ActionTicketSync     = "ticket.sync_status"

	// Aksi jenis tiket
	ActionTicketTypeCreate = "ticket_type.create"
	ActionTicketTypeUpdate = "ticket_type.update"
	ActionTicketTypeDelete = "ticket_type.delete"
	ActionTicketTypeStatus = "ticket_type.status_change"

	// Aksi pengguna
	ActionUserLogin    = "user.login"
	ActionUserRegister = "user.register"
	ActionUserProfile  = "user.profile_update"
)

const (
	StatusSuccess = "SUCCESS"
	StatusFailed  = "FAILED"
	StatusError   = "ERROR"
)
