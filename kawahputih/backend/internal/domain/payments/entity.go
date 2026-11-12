package payments

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

// Payment mencatat sisi uang dari booking. Satu booking punya satu payment aktif; percobaan FAILED bisa dicoba ulang di baris yang sama sampai jendela pembayaran habis. Field Midtrans opsional.
type Payment struct {
	Base
	BookingID        string     `gorm:"type:char(36);uniqueIndex;not null" json:"booking_id"`
	Method           string     `gorm:"type:varchar(20);not null;default:QRIS" json:"method"`
	Reference        string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"reference"`
	Amount           int64      `gorm:"not null" json:"amount"`
	Status           string     `gorm:"type:varchar(20);index;not null;default:PENDING" json:"status"`
	MidtransOrderID  string     `gorm:"type:varchar(100);uniqueIndex" json:"midtrans_order_id,omitempty"`
	MidtransToken    string     `gorm:"type:varchar(255)" json:"midtrans_token,omitempty"`
	MidtransRedirect string     `gorm:"type:varchar(500)" json:"midtrans_redirect_url,omitempty"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	ExpiredAt        time.Time  `gorm:"not null" json:"expired_at"`
}

func (Payment) TableName() string { return "payments" }

// Metode pembayaran untuk pengunjung. QRIS adalah default saat booking dibuat; pengunjung boleh beralih ke yang lain sebelum membayar.
const (
	MethodQRIS      = "QRIS"
	MethodGoPay     = "GOPAY"
	MethodOVO       = "OVO"
	MethodDANA      = "DANA"
	MethodShopeePay = "SHOPEEPAY"
	MethodBCAVA     = "BCA_VA"
	MethodBNIVA     = "BNI_VA"
	MethodMandiriVA = "MANDIRI_VA"
	MethodBRIVA     = "BRI_VA"
	MethodCard      = "CARD"
	MethodAlfamart  = "ALFAMART"
	MethodIndomaret = "INDOMARET"
)

// Methods adalah himpunan yang diterima untuk validasi request, disinkronkan dengan konstanta di atas dan daftar metode frontend.
var Methods = []string{
	MethodQRIS,
	MethodGoPay, MethodOVO, MethodDANA, MethodShopeePay,
	MethodBCAVA, MethodBNIVA, MethodMandiriVA, MethodBRIVA,
	MethodCard,
	MethodAlfamart, MethodIndomaret,
}

// ValidMethod menyatakan apakah kode metode didukung.
func ValidMethod(method string) bool {
	for _, m := range Methods {
		if m == method {
			return true
		}
	}
	return false
}

const (
	StatusPending    = "PENDING"
	StatusProcessing = "PROCESSING"
	StatusPaid       = "PAID"
	StatusFailed     = "FAILED"
	StatusExpired    = "EXPIRED"
	StatusRefunded   = "REFUNDED"
)

// PaymentWindow adalah berapa lama pembayaran pending tetap bisa dibayar. Cocok dengan sweep booking kedaluwarsa scheduler (30 menit).
const PaymentWindow = 30 * time.Minute
