package auth

import "github.com/KawahPutih/backend/pkg/apperror"

var (
	ErrUserNotFound         = apperror.NotFound("Pengguna tidak ditemukan.").WithCode("USER_001")
	ErrEmailAlreadyExists   = apperror.Conflict("Data sudah terdaftar.")
	ErrInvalidCredentials   = apperror.Unauthorized("Email atau password tidak valid.").WithCode("AUTH_001")
	ErrAccountDeactivated   = apperror.Forbidden("Akun Anda tidak aktif. Hubungi administrator.")
	// AUTH-06: login wajib email terverifikasi. Pesan mengarahkan pengguna ke
	// kotak masuk agar alur "kirim ulang verifikasi" bisa lanjut dari halaman login.
	ErrEmailNotVerified     = apperror.Forbidden("Email belum diverifikasi. Cek kotak masuk Anda untuk link verifikasi atau minta kirim ulang.").WithCode("AUTH_004")
	ErrInvalidToken         = apperror.Unauthorized("Sesi tidak valid atau sudah kedaluwarsa. Silakan login kembali.").WithCode("AUTH_002")
	ErrTokenRevoked         = apperror.Unauthorized("Sesi sudah berakhir. Silakan login kembali.").WithCode("AUTH_002")
	ErrOAuthNotConfigured   = apperror.Validation("Login Google belum tersedia.")
	ErrOAuthEmailUnverified = apperror.Unauthorized("Email akun Google belum terverifikasi.")

	ErrCurrentPasswordWrong = apperror.Unauthorized("Password saat ini tidak sesuai.").WithCode("AUTH_003")
	ErrPasswordNotSet       = apperror.Validation("Akun ini masuk lewat Google dan belum punya password. Gunakan 'Lupa Password' untuk membuatnya.")
	ErrPasswordReused       = apperror.Validation("Password baru harus berbeda dari password saat ini.")
)
