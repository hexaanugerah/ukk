package role

import "github.com/KawahPutih/backend/pkg/apperror"

var (
	ErrRoleNotFound = apperror.NotFound("Data role tidak ditemukan.")
)
