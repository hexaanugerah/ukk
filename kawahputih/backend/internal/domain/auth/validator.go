package auth

import (
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/go-playground/validator/v10"
)

// validate bersifat per-package (satu instance per domain), bukan global bersama semua domain, sesuai prinsip tiap domain independen.
var validate = validator.New()

func Validate(s interface{}) []response.FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var out []response.FieldError
	for _, fe := range err.(validator.ValidationErrors) {
		out = append(out, response.FieldError{Field: fe.Field(), Message: friendlyMessage(fe)})
	}
	return out
}

// ValidatePasswordStrength menegakkan kebijakan password di sisi server agar tidak bisa dilewati dengan memanggil API langsung; aturan frontend hanya kemudahan, bukan batas keamanan. Kebijakan: 8-72 karakter dengan minimal satu huruf besar, satu huruf kecil, dan satu angka. Karakter khusus boleh tapi tidak wajib.
func ValidatePasswordStrength(field, password string) []response.FieldError {
	// bcrypt memotong input di 72 BYTE, bukan 72 karakter. Password multibyte
	// (mis. aksen/emoji) yang lolos batas karakter bisa ditolak bcrypt dan
	// berujung 500, jadi batas atas diperiksa dalam byte.
	if len(password) < 8 || len([]byte(password)) > 72 {
		return []response.FieldError{{Field: field, Message: field + " must be between 8 and 72 characters"}}
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return []response.FieldError{{Field: field, Message: field + " must contain uppercase, lowercase, and a number"}}
	}
	return nil
}

func friendlyMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "email":
		return fe.Field() + " must be a valid email address"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters"
	default:
		return fe.Field() + " is invalid"
	}
}
