package public

import "github.com/KawahPutih/backend/pkg/apperror"

var (
	ErrNotFound        = apperror.NotFound("Data tidak ditemukan.")
	ErrInvalidCategory = apperror.Validation("Kategori tidak valid.")
)
