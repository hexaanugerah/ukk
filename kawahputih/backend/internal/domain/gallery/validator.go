package gallery

import (
	"github.com/KawahPutih/backend/pkg/response"
)

// Validate memvalidasi struktur request galeri dan mengembalikan daftar
// kesalahan per bidang (kosong jika valid).
func Validate(s interface{}) []response.FieldError {
	return validateStruct(s)
}
