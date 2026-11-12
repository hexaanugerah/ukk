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
	EWalletNumber    string     `gorm:"type:varchar(50)" json:"ewallet_number,omitempty"`
}

func (Payment) TableName() string { return "payments" }

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

	EwalletGoPay     = "GOPAY"
	EwalletOVO       = "OVO"
	EwalletDANA      = "DANA"
	EwalletShopeePay = "SHOPEEPAY"

	EwalletBCAVA     = "BCA_VA"
	EwalletBNIVA     = "BNI_VA"
	EwalletMandiriVA = "MANDIRI_VA"
	EwalletBRIVA     = "BRI_VA"
)

var Methods = []string{
	MethodQRIS,
	MethodGoPay, MethodOVO, MethodDANA, MethodShopeePay,
	MethodBCAVA, MethodBNIVA, MethodMandiriVA, MethodBRIVA,
	MethodCard,
	MethodAlfamart, MethodIndomaret,
}

var EwalletNumbers = []string{
	EwalletGoPay, EwalletOVO, EwalletDANA, EwalletShopeePay,
	EwalletBCAVA, EwalletBNIVA, EwalletMandiriVA, EwalletBRIVA,
}

func ValidMethod(m string) bool {
	for _, v := range Methods {
		if v == m {
			return true
		}
	}
	return false
}

func ValidEWalletNumber(v string) bool {
	for _, n := range EwalletNumbers {
		if n == v {
			return true
		}
	}
	return false
}

func EwalletNumberFromMethod(method string) string {
	switch method {
	case MethodGoPay:
		return EwalletGoPay
	case MethodOVO:
		return EwalletOVO
	case MethodDANA:
		return EwalletDANA
	case MethodShopeePay:
		return EwalletShopeePay
	case MethodBCAVA:
		return EwalletBCAVA
	case MethodBNIVA:
		return EwalletBNIVA
	case MethodMandiriVA:
		return EwalletMandiriVA
	case MethodBRIVA:
		return EwalletBRIVA
	default:
		return ""
	}
}

const (
	StatusPending    = "PENDING"
	StatusProcessing = "PROCESSING"
	StatusPaid       = "PAID"
	StatusFailed     = "FAILED"
	StatusExpired    = "EXPIRED"
	StatusRefunded   = "REFUNDED"
)

const PaymentWindow = 30 * time.Minute
