package tickettype

import (
	"fmt"
	"strings"

	"github.com/KawahPutih/backend/pkg/apperror"
)

var (
	ErrNotFound        = apperror.NotFound("Data tiket tidak ditemukan.")
	ErrCodeExists      = apperror.Conflict("Data sudah terdaftar.")
	ErrInvalidCategory = apperror.Validation("Kategori tidak valid. Pilih salah satu: " + strings.Join(Categories, ", "))
	ErrHasBookings     = apperror.Conflict("Jenis tiket tidak dapat dihapus karena masih memiliki booking.")
)

func invalidPrice() error { return apperror.Validation("Harga tiket harus lebih dari nol.") }
func invalidCapacity() error {
	return apperror.Validation("Kapasitas tiket harus lebih dari nol.")
}
func invalidStatus() error {
	return apperror.Validation(fmt.Sprintf("Status harus %s atau %s.", StatusActive, StatusInactive))
}
