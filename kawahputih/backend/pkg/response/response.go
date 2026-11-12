// Package response mendefinisikan satu envelope JSON yang dikembalikan setiap endpoint.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Nama field Meta (page/limit/total/total_pages) cocok persis dengan spesifikasi paginasi. Ini rename dari bentuk lama, jadi tipe ApiMeta frontend dan semua pembaca meta.* harus ikut diperbarui.
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// FieldError adalah satu entri pada respons validasi 422 dengan bentuk field dan message.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Success: true, Message: message, Data: data})
}

// Success adalah alias OK untuk kompatibilitas mundur
func Success(c *gin.Context, message string, data interface{}) {
	OK(c, message, data)
}

func OKWithMeta(c *gin.Context, message string, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Envelope{Success: true, Message: message, Data: data, Meta: meta})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Message: message, Data: data})
}

func Fail(c *gin.Context, status int, message string, errs interface{}) {
	c.JSON(status, Envelope{Success: false, Message: message, Errors: errs})
}

func BadRequest(c *gin.Context, message string, errs interface{}) {
	Fail(c, http.StatusBadRequest, message, errs)
}

// ValidationFailed adalah respons 422 khusus untuk error validasi request. Dibedakan dari BadRequest (400): body JSON rusak (tidak bisa diparse) adalah 400, sedangkan request valid yang gagal validasi field adalah 422.
func ValidationFailed(c *gin.Context, errs []FieldError) {
	Fail(c, http.StatusUnprocessableEntity, "Mohon lengkapi data terlebih dahulu.", errs)
}

func Unauthorized(c *gin.Context, message string) { Fail(c, http.StatusUnauthorized, message, nil) }
func Forbidden(c *gin.Context, message string)    { Fail(c, http.StatusForbidden, message, nil) }
func NotFound(c *gin.Context, message string)     { Fail(c, http.StatusNotFound, message, nil) }
func Conflict(c *gin.Context, message string)     { Fail(c, http.StatusConflict, message, nil) }
func TooManyRequests(c *gin.Context, message string) {
	Fail(c, http.StatusTooManyRequests, message, nil)
}
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = "Terjadi kesalahan. Silakan coba kembali."
	}
	Fail(c, http.StatusInternalServerError, message, nil)
}

// BuildMeta menghitung total_pages dari pasangan limit/total. page dan limit di sini adalah nilai yang benar-benar dipakai query, bukan input mentah.
func BuildMeta(page, limit int, total int64) *Meta {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	return &Meta{Page: page, Limit: limit, Total: total, TotalPages: totalPages}
}
