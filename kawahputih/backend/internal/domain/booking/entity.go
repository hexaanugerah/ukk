package booking

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

// Booking adalah pesanan multi-item: satu tanggal kunjungan, satu atau lebih baris tiket, total dihitung server. Semua nominal bilangan bulat rupiah.
// Transisi status (ditegakkan di service.go):
// PENDING --bayar--> PAID --scan petugas--> USED
// PENDING --batal--> CANCELLED
// PENDING --kedaluwarsa/sweeper--> EXPIRED
type Booking struct {
	Base
	BookingCode   string     `gorm:"type:varchar(30);uniqueIndex;not null" json:"booking_code"`
	UserID        string     `gorm:"type:char(36);index;not null" json:"user_id"`
	VisitorName   string     `gorm:"type:varchar(150);not null" json:"visitor_name"`
	Email         string     `gorm:"type:varchar(150);not null" json:"email"`
	Phone         string     `gorm:"type:varchar(20);not null" json:"phone"`
	VisitDate     string     `gorm:"type:varchar(10);index;not null" json:"visit_date"` // "2006-01-02" local date
	Note          string     `gorm:"type:varchar(500)" json:"note"`
	Status        string     `gorm:"type:varchar(20);index;not null;default:PENDING" json:"status"`
	Subtotal      int64      `gorm:"not null;default:0" json:"subtotal"`
	Discount      int64      `gorm:"not null;default:0" json:"discount"`
	ServiceFee    int64      `gorm:"not null;default:0" json:"service_fee"`
	Tax           int64      `gorm:"not null;default:0" json:"tax"`
	Total         int64      `gorm:"not null;default:0" json:"total"`
	TotalTickets  int        `gorm:"not null;default:0" json:"total_tickets"`
	CheckedInAt   *time.Time `json:"checked_in_at,omitempty"`
	CheckedInGate string     `gorm:"type:varchar(100)" json:"checked_in_gate,omitempty"`

	Items []BookingItem `gorm:"foreignKey:BookingID" json:"items"`
}

func (Booking) TableName() string { return "bookings" }

// BookingItem adalah satu baris pesanan: jenis tiket, harga satuan (snapshot saat booking sehingga perubahan harga nanti tidak mengubah pesanan lama), dan jumlah.
type BookingItem struct {
	Base
	BookingID  string `gorm:"type:char(36);index;not null" json:"booking_id"`
	TicketID   string `gorm:"type:char(36);index;not null" json:"ticket_id"`
	TicketName string `gorm:"type:varchar(150);not null" json:"ticket_name"`
	Price      int64  `gorm:"not null" json:"price"`
	Quantity   int    `gorm:"not null" json:"quantity"`
}

func (BookingItem) TableName() string { return "booking_items" }

const (
	StatusPending   = "PENDING"
	StatusPaid      = "PAID"
	StatusConfirmed = "CONFIRMED"
	StatusUsed      = "USED"
	StatusCancelled = "CANCELLED"
	StatusExpired   = "EXPIRED"
)

// RevenueStatuses adalah status booking yang dihitung sebagai stok terjual dan pendapatan. Dipakai bersama analytics/laporan.
var RevenueStatuses = []string{StatusPaid, StatusConfirmed, StatusUsed}
