package visitorrule

import "github.com/KawahPutih/backend/pkg/apperror"

var (
	ErrNotFound = apperror.NotFound("Data aturan pengunjung tidak ditemukan.")
)
