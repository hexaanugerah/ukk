package user

import "github.com/KawahPutih/backend/pkg/apperror"

// USER_001 sama seperti auth.ErrUserNotFound, karena keduanya mewakili kondisi nyata yang sama.
var ErrUserNotFound = apperror.NotFound("Pengguna tidak ditemukan.").WithCode("USER_001")
