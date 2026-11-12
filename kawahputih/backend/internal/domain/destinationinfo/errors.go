package destinationinfo

import "github.com/KawahPutih/backend/pkg/apperror"

var (
	ErrNotFound      = apperror.NotFound("Data informasi destinasi tidak ditemukan.")
	ErrTitleExists   = apperror.Conflict("Data sudah terdaftar.")
	ErrInvalidStatus = apperror.Validation("Perubahan status tidak valid.")
	ErrCannotPublish = apperror.Validation("Konten belum lengkap untuk dipublikasikan.")
)
