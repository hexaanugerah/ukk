package auth

import (
	"strings"
	"testing"
)

// TestValidatePasswordStrength mengunci kebijakan password: 8-72 karakter dengan huruf besar + kecil + angka, dan tidak wajib karakter khusus.
func TestValidatePasswordStrength(t *testing.T) {
	valid := []string{
		"AdminKawah2026",
		"pengunjung2026A",
		"ManagerKawah2026",
		"Abcdefg1",         // exactly 8 chars
		"Abcdefg1WithSym!", // symbols allowed but not required
	}
	for _, pw := range valid {
		if errs := ValidatePasswordStrength("password", pw); len(errs) != 0 {
			t.Errorf("%q: expected valid, got %v", pw, errs)
		}
	}

	invalid := map[string]string{
		"Ab1":       "too short",
		"abcdefgh1": "no uppercase",
		"ABCDEFGH1": "no lowercase",
		"Abcdefgh":  "no digit",
		// 48 karakter tetapi 88 byte: melewati batas 72 byte bcrypt.
		"Abcdefg1" + strings.Repeat("é", 40): "over 72 bytes",
	}
	for pw, reason := range invalid {
		if errs := ValidatePasswordStrength("password", pw); len(errs) == 0 {
			t.Errorf("%q: expected invalid (%s), got valid", pw, reason)
		}
	}
}

// TestValidatePasswordStrengthFieldName memastikan error dilaporkan pada field yang diminta pemanggil (misalnya new_password di endpoint reset).
func TestValidatePasswordStrengthFieldName(t *testing.T) {
	errs := ValidatePasswordStrength("new_password", "short")
	if len(errs) == 0 {
		t.Fatal("expected an error for a short password")
	}
	if errs[0].Field != "new_password" {
		t.Errorf("field = %q, want new_password", errs[0].Field)
	}
}
