package apperror

import (
	"net/http"

	"github.com/KawahPutih/backend/pkg/logger"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// errorBody bentuknya sama dengan response.Envelope tapi menambah field code opsional; tipe lokal, karena sebagian besar respons tidak membawa kode.
type errorBody struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// WriteHTTP dipanggil tiap handler.go, satu tempat memetakan Kind ke status HTTP agar semua domain merespons sama untuk kelas error yang sama. Error validasi bisnis dari service/usecase masuk sebagai KindValidation dengan bentuk 422 pesan biasa; bentuk 422 per-field dipakai handler untuk validasi binding.
func WriteHTTP(c *gin.Context, err error) {
	ae, ok := As(err)
	if !ok {
		serverError(c, "", err)
		c.JSON(http.StatusInternalServerError, errorBody{Success: false, Message: "Terjadi kesalahan. Silakan coba kembali."})
		return
	}

	// Detail error hanya masuk log server; klien selalu menerima pesan yang aman.
	if ae.Kind == KindInternal {
		serverError(c, ae.Code, ae.Err)
	}

	status := statusFor(ae.Kind)
	if ae.Code != "" {
		c.JSON(status, errorBody{Success: false, Message: ae.Message, Code: ae.Code})
		return
	}
	// Tanpa kode katalog, jadi jatuh ke envelope biasa agar helper response.* tetap jadi acuan.
	switch ae.Kind {
	case KindValidation:
		c.JSON(http.StatusUnprocessableEntity, errorBody{Success: false, Message: ae.Message})
	case KindNotFound:
		response.NotFound(c, ae.Message)
	case KindConflict:
		response.Conflict(c, ae.Message)
	case KindUnauthorized:
		response.Unauthorized(c, ae.Message)
	case KindForbidden:
		response.Forbidden(c, ae.Message)
	case KindRateLimited:
		response.TooManyRequests(c, ae.Message)
	default:
		c.JSON(http.StatusInternalServerError, errorBody{Success: false, Message: "Terjadi kesalahan. Silakan coba kembali."})
	}
}

// serverError mencatat penyebab internal ke log server tanpa mengirim detailnya ke klien.
func serverError(c *gin.Context, code string, err error) {
	fields := []zap.Field{
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("code", code),
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	logger.L().Error("internal error", fields...)
}

func statusFor(kind Kind) int {
	switch kind {
	case KindValidation:
		return http.StatusUnprocessableEntity
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
