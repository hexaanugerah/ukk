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
	Provider         string     `gorm:"type:varchar(30);index;not null;default:midtrans" json:"provider"`
	Method           string     `gorm:"type:varchar(20);not null;default:QRIS" json:"method"`
	Reference        string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"reference"`
	Amount           int64      `gorm:"not null" json:"amount"`
	Status           string     `gorm:"type:varchar(20);index;not null;default:PENDING" json:"status"`
	SettlementStatus string     `gorm:"type:varchar(20);index;not null;default:UNSETTLED" json:"settlement_status"`
	MidtransOrderID  string     `gorm:"type:varchar(100);uniqueIndex" json:"midtrans_order_id,omitempty"`
	MidtransToken    string     `gorm:"type:varchar(255)" json:"midtrans_token,omitempty"`
	MidtransRedirect string     `gorm:"type:varchar(500)" json:"midtrans_redirect_url,omitempty"`
	QrisString       string     `gorm:"type:text" json:"-"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	ExpiredAt        time.Time  `gorm:"not null" json:"expired_at"`
	EWalletNumber    string     `gorm:"type:varchar(50)" json:"ewallet_number,omitempty"`
	// ProofURL adalah bukti pembayaran yang diunggah pengunjung. Wajib ada
	// sebelum admin boleh memverifikasi pembayaran manual menjadi PAID.
	ProofURL        string     `gorm:"type:varchar(500)" json:"proof_url,omitempty"`
	ProofUploadedAt *time.Time `json:"proof_uploaded_at,omitempty"`
}

// PaymentSetting menyimpan konfigurasi payment gateway. Kredensial rahasia
// (server key / access token) TIDAK disimpan di sini — hanya di environment.
type PaymentSetting struct {
	Base
	Provider                string     `gorm:"type:varchar(30);not null;default:midtrans" json:"provider"`
	Environment             string     `gorm:"type:varchar(20);not null;default:sandbox" json:"environment"`
	Enabled                 bool       `gorm:"not null;default:true" json:"enabled"`
	MerchantName            string     `gorm:"type:varchar(150)" json:"merchant_name"`
	SettlementBank          string     `gorm:"type:varchar(50)" json:"settlement_bank"`
	SettlementAccountMasked string     `gorm:"type:varchar(50)" json:"settlement_account_masked"`
	SettlementStatus        string     `gorm:"type:varchar(20);not null;default:PENDING" json:"settlement_status"`
	LastSyncedAt            *time.Time `json:"last_synced_at,omitempty"`
}

func (PaymentSetting) TableName() string { return "payment_settings" }

// PaymentMethodConfig adalah kanal pembayaran yang dapat diaktifkan admin.
// Aktif di sini belum berarti aktif di akun merchant — validasi tetap lewat provider.
type PaymentMethodConfig struct {
	Base
	Provider     string `gorm:"type:varchar(30);not null;default:midtrans" json:"provider"`
	MethodCode   string `gorm:"type:varchar(20);uniqueIndex;not null" json:"method_code"`
	MethodName   string `gorm:"type:varchar(100);not null" json:"method_name"`
	Enabled      bool   `gorm:"not null;default:true" json:"enabled"`
	DisplayOrder int    `gorm:"default:0" json:"display_order"`
}

func (PaymentMethodConfig) TableName() string { return "payment_methods" }

// PaymentWebhookLog mencatat setiap notifikasi provider untuk mencegah
// pemrosesan ganda (idempotency) dan menyimpan jejak audit.
type PaymentWebhookLog struct {
	Base
	Provider         string     `gorm:"type:varchar(30);not null;default:midtrans" json:"provider"`
	TransactionID    string     `gorm:"type:varchar(100);index" json:"transaction_id"`
	NotificationID   string     `gorm:"type:varchar(150);uniqueIndex;not null" json:"notification_id"`
	EventType        string     `gorm:"type:varchar(50)" json:"event_type"`
	ProcessingStatus string     `gorm:"type:varchar(20);not null;default:RECEIVED" json:"processing_status"`
	ReceivedAt       time.Time  `json:"received_at"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
}

func (PaymentWebhookLog) TableName() string { return "payment_webhook_logs" }

func (Payment) TableName() string { return "payments" }

// PaymentDestination adalah tujuan pembayaran manual (rekening bank, nomor
// e-wallet, kode retail) yang dikelola oleh admin.
type PaymentDestination struct {
	Base
	MethodCode    string `gorm:"type:varchar(20);index;not null" json:"method_code"`
	ProviderName  string `gorm:"type:varchar(100);not null" json:"provider_name"`
	AccountNumber string `gorm:"type:varchar(100)" json:"account_number"`
	AccountName   string `gorm:"type:varchar(100)" json:"account_name"`
	Active        bool   `gorm:"not null;default:true" json:"active"`
	SortOrder     int    `gorm:"default:0" json:"sort_order"`
}

func (PaymentDestination) TableName() string { return "payment_destinations" }

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

// Status settlement dipisahkan dari status transaksi: pembayaran berhasil
// belum tentu dananya sudah cair ke rekening merchant.
const (
	SettlementUnsettled = "UNSETTLED"
	SettlementSettled   = "SETTLED"
	SettlementRefunded  = "REFUNDED"
)

const ProviderMidtrans = "midtrans"

// WebhookProcessing* adalah status pemrosesan log notifikasi provider.
const (
	WebhookReceived  = "RECEIVED"
	WebhookProcessed = "PROCESSED"
	WebhookFailed    = "FAILED"
	WebhookDuplicate = "DUPLICATE"
)

const PaymentWindow = 30 * time.Minute
