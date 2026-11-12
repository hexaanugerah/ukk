package booking

import (
	"github.com/KawahPutih/backend/pkg/apperror"
)

var (
	ErrNotFound        = apperror.NotFound("Data booking tidak ditemukan.").WithCode("BOOK_001")
	ErrNotOwner        = apperror.Forbidden("Anda tidak memiliki akses.").WithCode("BOOK_002")
	ErrNoItems         = apperror.Validation("Mohon pilih minimal satu tiket.").WithCode("BOOK_003")
	ErrInvalidDate     = apperror.Validation("Tanggal kunjungan tidak valid.").WithCode("BOOK_004")
	ErrInvalidQuantity = apperror.Validation("Jumlah tiket tidak valid.").WithCode("BOOK_005")
	ErrUnavailable     = apperror.Conflict("Tiket pada tanggal tersebut tidak tersedia.").WithCode("BOOK_006")
	ErrCannotCancel    = apperror.Conflict("Booking ini tidak dapat dibatalkan.").WithCode("BOOK_007")
	ErrInvalidStatus   = apperror.Validation("Status booking tidak valid.").WithCode("BOOK_008")
)

func unknownTicket(id string) error {
	_ = id // detail id tidak ditampilkan ke pengguna agar pesan tetap ramah
	return apperror.NotFound("Data tiket tidak ditemukan.").WithCode("BOOK_009")
}
