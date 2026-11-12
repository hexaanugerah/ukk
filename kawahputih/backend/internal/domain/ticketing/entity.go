package ticketing

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

// Ticket adalah e-tiket yang diterbitkan, satu per booking berbayar. QRData adalah identitas unik yang dikodekan ke QR; Status mengikuti siklus hidup booking (VALID sampai di-scan/dibatalkan/kedaluwarsa).
type Ticket struct {
	Base
	BookingID     string     `gorm:"type:char(36);uniqueIndex;not null" json:"booking_id"`
	BookingCode   string     `gorm:"type:varchar(30);index" json:"booking_code"`
	VisitorName   string     `gorm:"type:varchar(150);not null" json:"visitor_name"`
	Email         string     `gorm:"type:varchar(150);not null" json:"email"`
	VisitDate     string     `gorm:"type:varchar(10);index;not null" json:"visit_date"`
	TicketName    string     `gorm:"type:varchar(300);not null" json:"ticket_name"`
	Quantity      int        `gorm:"not null" json:"quantity"`
	Total         int64      `gorm:"not null" json:"total"`
	Status        string     `gorm:"type:varchar(20);index;not null;default:VALID" json:"status"`
	GeneratedAt   time.Time  `gorm:"not null" json:"generated_at"`
	CheckedInAt   *time.Time `json:"checked_in_at,omitempty"`
	CheckedInGate string     `gorm:"type:varchar(100)" json:"checked_in_gate,omitempty"`
	QRData        string     `gorm:"type:varchar(120);uniqueIndex;not null" json:"qr_data"`
}

func (Ticket) TableName() string { return "tickets" }

// ScanRecord adalah satu percobaan verifikasi, berhasil atau tidak. Ini riwayat verifikasi petugas sekaligus sumber jumlah pengunjung.
type ScanRecord struct {
	Base
	TicketID    string    `gorm:"type:char(36);index;not null" json:"ticket_id"`
	BookingID   string    `gorm:"type:char(36);index;not null" json:"booking_id"`
	VisitorName string    `gorm:"type:varchar(150);not null" json:"visitor_name"`
	Gate        string    `gorm:"type:varchar(100)" json:"gate"`
	PetugasID   string    `gorm:"type:char(36);index" json:"petugas_id"`
	PetugasName string    `gorm:"type:varchar(150)" json:"petugas"`
	Status      string    `gorm:"type:varchar(20);index;not null" json:"status"`
	ScannedAt   time.Time `gorm:"index;not null" json:"timestamp"`
}

func (ScanRecord) TableName() string { return "scan_records" }

const (
	TicketValid     = "VALID"
	TicketUsed      = "USED"
	TicketCancelled = "CANCELLED"
	TicketExpired   = "EXPIRED"

	ScanValid     = "VALID"
	ScanUsed      = "USED"
	ScanInvalid   = "INVALID"
	ScanExpired   = "EXPIRED"
	ScanCancelled = "CANCELLED"
)
