package validation

import (
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Instance Validator dengan validasi kustom
var validate = validator.New()

func init() {
	// Daftarkan validator kustom
	_ = validate.RegisterValidation("phone_id", validateIndonesianPhone)
	_ = validate.RegisterValidation("date_future", validateFutureDate)
	_ = validate.RegisterValidation("date_range", validateDateRange)
	_ = validate.RegisterValidation("no_sql_injection", validateNoSQLInjection)
	_ = validate.RegisterValidation("no_xss", validateNoXSS)
	_ = validate.RegisterValidation("secure_string", validateSecureString)
}

// ValidateStruct memvalidasi struct dan mengembalikan error per field
func ValidateStruct(s interface{}) []FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var out []FieldError
	for _, fe := range err.(validator.ValidationErrors) {
		out = append(out, FieldError{
			Field:   fe.Field(),
			Message: translateValidationError(fe),
		})
	}
	return out
}

// ValidateVar memvalidasi satu variabel dengan tag
func ValidateVar(field interface{}, tag string) []FieldError {
	err := validate.Var(field, tag)
	if err == nil {
		return nil
	}

	var out []FieldError
	for _, fe := range err.(validator.ValidationErrors) {
		out = append(out, FieldError{
			Field:   fe.Field(),
			Message: translateValidationError(fe),
		})
	}
	return out
}

// FieldError merepresentasikan error validasi untuk field tertentu
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// translateValidationError mengubah error validator menjadi pesan yang mudah dipahami
func translateValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters"
	case "email":
		return fe.Field() + " must be a valid email address"
	case "gt":
		return fe.Field() + " must be greater than " + fe.Param()
	case "gte":
		return fe.Field() + " must be greater than or equal to " + fe.Param()
	case "lt":
		return fe.Field() + " must be less than " + fe.Param()
	case "lte":
		return fe.Field() + " must be less than or equal to " + fe.Param()
	case "oneof":
		return fe.Field() + " must be one of: " + fe.Param()
	case "phone_id":
		return fe.Field() + " must be a valid Indonesian phone number"
	case "date_future":
		return fe.Field() + " must be a future date"
	case "date_range":
		return fe.Field() + " must be within allowed date range"
	case "no_sql_injection":
		return fe.Field() + " contains potentially dangerous characters"
	case "no_xss":
		return fe.Field() + " contains potentially dangerous characters"
	case "secure_string":
		return fe.Field() + " contains invalid characters"
	case "alphanum":
		return fe.Field() + " must contain only alphanumeric characters"
	case "alpha":
		return fe.Field() + " must contain only letters"
	case "numeric":
		return fe.Field() + " must contain only numbers"
	case "len":
		return fe.Field() + " must be exactly " + fe.Param() + " characters"
	case "uuid":
		return fe.Field() + " must be a valid UUID"
	case "url":
		return fe.Field() + " must be a valid URL"
	default:
		return fe.Field() + " is invalid (" + fe.Tag() + ")"
	}
}

// validateIndonesianPhone memvalidasi format nomor telepon Indonesia
func validateIndonesianPhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	cleaned := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")

	// Ubah +62 menjadi 0
	if strings.HasPrefix(cleaned, "+62") {
		cleaned = "0" + cleaned[3:]
	} else if strings.HasPrefix(cleaned, "62") {
		cleaned = "0" + cleaned[2:]
	}

	// Harus diawali 08 dan 10-13 digit
	if !strings.HasPrefix(cleaned, "08") {
		return false
	}
	return len(cleaned) >= 10 && len(cleaned) <= 13
}

// validateFutureDate memvalidasi bahwa string tanggal ada di masa depan
func validateFutureDate(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}
	today := time.Now().Truncate(24 * time.Hour)
	return t.After(today) || t.Equal(today)
}

// validateDateRange memvalidasi tanggal berada dalam jendela pemesanan yang diizinkan
func validateDateRange(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}
	today := time.Now().Truncate(24 * time.Hour)
	maxDate := today.AddDate(0, 0, 60) // Max 60 days ahead
	return !t.Before(today) && !t.After(maxDate)
}

// validateNoSQLInjection mengecek pola potensi SQL injection
func validateNoSQLInjection(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())
	sqlPatterns := []string{
		"union select", "drop table", "delete from", "insert into",
		"update set", "alter table", "create table", "truncate table",
		"exec(", "execute(", "sp_", "xp_", "--", ";--", "/*", "*/",
		"char(", "nchar(", "varchar(", "nvarchar(", "cast(", "convert(",
		"waitfor delay", "benchmark(", "sleep(", "pg_sleep",
	}

	for _, pattern := range sqlPatterns {
		if strings.Contains(value, pattern) {
			return false
		}
	}
	return true
}

// validateNoXSS mengecek pola potensi XSS
func validateNoXSS(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "onerror=", "onload=",
		"onclick=", "onmouseover=", "onfocus=", "onblur=",
		"onchange=", "onsubmit=", "onkeydown=", "onkeyup=",
		"onkeypress=", "<iframe", "<object", "<embed", "<form",
		"<input", "<img", "<svg", "expression(", "vbscript:",
		"data:text/html", "alert(", "confirm(", "prompt(",
	}

	for _, pattern := range xssPatterns {
		if strings.Contains(value, pattern) {
			return false
		}
	}
	return true
}

// validateSecureString mengecek karakter kontrol dan konten berbahaya lain
func validateSecureString(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Cek null byte dan karakter kontrol (kecuali newline/tab)
	for _, r := range value {
		if r == 0 || (r < 32 && r != '\n' && r != '\t' && r != '\r') {
			return false
		}
	}
	return true
}

// SanitizeInput menghapus karakter berbahaya dari input
func SanitizeInput(input string) string {
	// Hapus null byte
	input = strings.ReplaceAll(input, "\x00", "")
	// Pangkas spasi
	input = strings.TrimSpace(input)
	// Batasi panjang
	if len(input) > 10000 {
		input = input[:10000]
	}
	return input
}

// SanitizeEmail menormalkan email
func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// SanitizePhone menormalkan nomor telepon Indonesia
func SanitizePhone(phone string) string {
	cleaned := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")
	if strings.HasPrefix(cleaned, "+62") {
		cleaned = "0" + cleaned[3:]
	} else if strings.HasPrefix(cleaned, "62") {
		cleaned = "0" + cleaned[2:]
	}
	return cleaned
}

// IsValidEmail memvalidasi format email (sesuai RFC 5322)
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$`)
	return emailRegex.MatchString(email)
}

// IsValidTicketCode memvalidasi format kode tiket
func IsValidTicketCode(code string) bool {
	// Kode tiket: alfanumerik, 3-20 karakter
	matched, _ := regexp.MatchString(`^[A-Z0-9]{3,20}$`, strings.ToUpper(code))
	return matched
}

// IsValidBookingID memvalidasi format ID booking
func IsValidBookingID(id string) bool {
	// Format: KP-YYYYMMDD-XXXXXX
	matched, _ := regexp.MatchString(`^KP-\d{8}-\d{6}$`, id)
	return matched
}

// IsValidPaymentID memvalidasi format ID pembayaran
func IsValidPaymentID(id string) bool {
	// Format: PAY-YYYYMMDD-XXXXXX
	matched, _ := regexp.MatchString(`^PAY-\d{8}-\d{6}$`, id)
	return matched
}

// IsValidTicketID memvalidasi format ID e-tiket
func IsValidTicketID(id string) bool {
	// Format: TKT-XXXXXXXXXX
	matched, _ := regexp.MatchString(`^TKT-[A-F0-9]{10}$`, id)
	return matched
}

// ValidationRules memuat aturan validasi umum
var ValidationRules = struct {
	Name     string
	Email    string
	Phone    string
	Password string
	Note     string
	Quantity string
	Date     string
}{
	Name:     "required,min=3,max=100,secure_string,no_xss",
	Email:    "required,email,max=150",
	Phone:    "required,phone_id",
	Password: "required,min=8,max=72",
	Note:     "max=500,secure_string,no_xss,no_sql_injection",
	Quantity: "required,gte=1,lte=20",
	Date:     "required,date_range",
}

// ValidateBookingRequest memvalidasi request pembuatan booking
func ValidateBookingRequest(visitDate string, visitorName, visitorEmail, visitorPhone string, items []BookingItemRequest) []FieldError {
	var errors []FieldError

	// Validasi tanggal kunjungan
	if errs := ValidateVar(visitDate, "date_range"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	// Validasi nama pengunjung
	if errs := ValidateVar(visitorName, "required,min=3,max=100,secure_string,no_xss"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	// Validasi email pengunjung
	if errs := ValidateVar(visitorEmail, "required,email,max=150"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	// Validasi telepon pengunjung
	if errs := ValidateVar(visitorPhone, "phone_id"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	// Validasi item
	if len(items) == 0 {
		errors = append(errors, FieldError{Field: "items", Message: "At least one ticket item is required"})
	} else if len(items) > 10 {
		errors = append(errors, FieldError{Field: "items", Message: "Maximum 10 different ticket types per booking"})
	}

	seenTickets := make(map[string]bool)
	for i, item := range items {
		if seenTickets[item.TicketID] {
			errors = append(errors, FieldError{Field: "items[" + string(rune(i+'0')) + "].ticket_id", Message: "Duplicate ticket ID in booking"})
		}
		seenTickets[item.TicketID] = true

		if errs := ValidateVar(item.TicketID, "required,uuid"); len(errs) > 0 {
			for _, e := range errs {
				errors = append(errors, FieldError{Field: "items[" + string(rune(i+'0')) + "]." + e.Field, Message: e.Message})
			}
		}
		if errs := ValidateVar(item.Quantity, "gte=1,lte=20"); len(errs) > 0 {
			for _, e := range errs {
				errors = append(errors, FieldError{Field: "items[" + string(rune(i+'0')) + "]." + e.Field, Message: e.Message})
			}
		}
	}

	return errors
}

// ValidateTicketCreation memvalidasi request pembuatan tiket
func ValidateTicketCreation(code, name, description string, category string, price int64, capacity int, status string) []FieldError {
	var errors []FieldError

	if errs := ValidateVar(code, "required,alphanum,min=3,max=20"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	if errs := ValidateVar(name, "required,min=3,max=100,secure_string,no_xss"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	if errs := ValidateVar(description, "max=1000,secure_string,no_xss,no_sql_injection"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	validCategories := []string{"ENTRANCE", "TRANSPORT", "SUNRISE", "PACKAGE", "ACTIVITY", "SERVICE"}
	validCategory := false
	for _, c := range validCategories {
		if c == category {
			validCategory = true
			break
		}
	}
	if !validCategory && category != "" {
		errors = append(errors, FieldError{Field: "category", Message: "Invalid ticket category"})
	}

	if price <= 0 {
		errors = append(errors, FieldError{Field: "price", Message: "Price must be greater than 0"})
	}

	if capacity <= 0 {
		errors = append(errors, FieldError{Field: "capacity", Message: "Capacity must be greater than 0"})
	}

	if capacity > 10000 {
		errors = append(errors, FieldError{Field: "capacity", Message: "Capacity cannot exceed 10000"})
	}

	validStatuses := []string{"AKTIF", "NONAKTIF"}
	validStatus := false
	for _, s := range validStatuses {
		if s == status {
			validStatus = true
			break
		}
	}
	if !validStatus {
		errors = append(errors, FieldError{Field: "status", Message: "Status must be AKTIF or NONAKTIF"})
	}

	return errors
}

// ValidatePaymentRequest memvalidasi request pemrosesan pembayaran
func ValidatePaymentRequest(bookingID, forced string) []FieldError {
	var errors []FieldError

	if errs := ValidateVar(bookingID, "required"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	if forced != "" {
		validOutcomes := []string{"SUCCESS", "FAILED", "EXPIRED"}
		valid := false
		for _, o := range validOutcomes {
			if o == forced {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, FieldError{Field: "forced", Message: "Forced outcome must be SUCCESS, FAILED, or EXPIRED"})
		}
	}

	return errors
}

// ValidateTicketVerification memvalidasi request verifikasi tiket
func ValidateTicketVerification(ticketID, gate string) []FieldError {
	var errors []FieldError

	if errs := ValidateVar(ticketID, "required"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	if gate != "" {
		if errs := ValidateVar(gate, "max=50,secure_string,no_xss"); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	return errors
}

// ValidateAuthRequest memvalidasi request autentikasi
func ValidateAuthRequest(email, password string) []FieldError {
	var errors []FieldError

	if errs := ValidateVar(email, "required,email,max=150"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	if errs := ValidateVar(password, "required,min=8,max=72"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	return errors
}

// ValidateRegistrationRequest memvalidasi request registrasi pengguna
func ValidateRegistrationRequest(name, email, phone, password, confirmPassword string) []FieldError {
	var errors []FieldError

	if errs := ValidateVar(name, "required,min=3,max=100,secure_string,no_xss"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	if errs := ValidateVar(email, "required,email,max=150"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	if phone != "" {
		if errs := ValidateVar(phone, "phone_id"); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	if errs := ValidateVar(password, "required,min=8,max=72"); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	if password != confirmPassword {
		errors = append(errors, FieldError{Field: "confirmPassword", Message: "Passwords do not match"})
	}

	return errors
}
