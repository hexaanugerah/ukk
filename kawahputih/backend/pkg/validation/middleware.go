package validation

import (
	"net/http"
	"time"

	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationMiddleware membuat middleware yang memvalidasi body request terhadap sebuah struct
func ValidationMiddleware[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T
		if err := c.ShouldBindJSON(&req); err != nil {
			var fieldErrors []response.FieldError

			// Coba ekstrak error validasi
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				for _, fe := range validationErrors {
					fieldErrors = append(fieldErrors, response.FieldError{
						Field:   fe.Field(),
						Message: translateValidationError(fe),
					})
				}
			} else {
				fieldErrors = append(fieldErrors, response.FieldError{
					Field:   "body",
					Message: "Invalid request body: " + err.Error(),
				})
			}

			response.ValidationFailed(c, fieldErrors)
			c.Abort()
			return
		}

		// Simpan request yang tervalidasi di context
		c.Set("validated_request", req)
		c.Next()
	}
}

// QueryValidationMiddleware membuat middleware yang memvalidasi parameter query
func QueryValidationMiddleware[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T
		if err := c.ShouldBindQuery(&req); err != nil {
			var fieldErrors []response.FieldError

			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				for _, fe := range validationErrors {
					fieldErrors = append(fieldErrors, response.FieldError{
						Field:   fe.Field(),
						Message: translateValidationError(fe),
					})
				}
			} else {
				fieldErrors = append(fieldErrors, response.FieldError{
					Field:   "query",
					Message: "Invalid query parameters: " + err.Error(),
				})
			}

			response.ValidationFailed(c, fieldErrors)
			c.Abort()
			return
		}

		c.Set("validated_query", req)
		c.Next()
	}
}

// ParamValidationMiddleware membuat middleware yang memvalidasi parameter URL
func ParamValidationMiddleware[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T
		if err := c.ShouldBindUri(&req); err != nil {
			var fieldErrors []response.FieldError

			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				for _, fe := range validationErrors {
					fieldErrors = append(fieldErrors, response.FieldError{
						Field:   fe.Field(),
						Message: translateValidationError(fe),
					})
				}
			} else {
				fieldErrors = append(fieldErrors, response.FieldError{
					Field:   "params",
					Message: "Invalid URL parameters: " + err.Error(),
				})
			}

			response.ValidationFailed(c, fieldErrors)
			c.Abort()
			return
		}

		c.Set("validated_params", req)
		c.Next()
	}
}

// GetValidatedRequest mengambil request tervalidasi dari context
func GetValidatedRequest[T any](c *gin.Context) (T, bool) {
	val, exists := c.Get("validated_request")
	if !exists {
		var zero T
		return zero, false
	}
	req, ok := val.(T)
	return req, ok
}

// GetValidatedQuery mengambil query tervalidasi dari context
func GetValidatedQuery[T any](c *gin.Context) (T, bool) {
	val, exists := c.Get("validated_query")
	if !exists {
		var zero T
		return zero, false
	}
	req, ok := val.(T)
	return req, ok
}

// GetValidatedParams mengambil params tervalidasi dari context
func GetValidatedParams[T any](c *gin.Context) (T, bool) {
	val, exists := c.Get("validated_params")
	if !exists {
		var zero T
		return zero, false
	}
	req, ok := val.(T)
	return req, ok
}

// Struct ValidationRequest untuk endpoint umum

// BookingCreateRequest untuk middleware validasi
type BookingCreateRequest struct {
	VisitDate string                `json:"visit_date" binding:"required,date_range"`
	Visitor   BookingVisitorRequest `json:"visitor" binding:"required"`
	Items     []BookingItemRequest  `json:"items" binding:"required,min=1,max=10"`
	Note      string                `json:"note,omitempty" binding:"max=500,secure_string,no_xss,no_sql_injection"`
}

type BookingVisitorRequest struct {
	Name  string `json:"name" binding:"required,min=3,max=100,secure_string,no_xss"`
	Email string `json:"email" binding:"required,email,max=150"`
	Phone string `json:"phone" binding:"required,phone_id"`
}

type BookingItemRequest struct {
	TicketID string `json:"ticket_id" binding:"required,uuid"`
	Quantity int    `json:"quantity" binding:"required,gte=1,lte=20"`
}

// TicketCreateRequest untuk middleware validasi
type TicketCreateRequest struct {
	Code        string `json:"code" binding:"required,alphanum,len=3|20"`
	Name        string `json:"name" binding:"required,min=3,max=100,secure_string,no_xss"`
	Description string `json:"description" binding:"max=1000,secure_string,no_xss,no_sql_injection"`
	Category    string `json:"category" binding:"required,oneof=ENTRANCE TRANSPORT SUNRISE PACKAGE ACTIVITY SERVICE"`
	Price       int64  `json:"price" binding:"required,gt=0"`
	Capacity    int    `json:"capacity" binding:"required,gt=0,lte=10000"`
	Status      string `json:"status" binding:"required,oneof=AKTIF NONAKTIF"`
	Notes       string `json:"notes,omitempty" binding:"max=500,secure_string,no_xss,no_sql_injection"`
}

// PaymentProcessRequest untuk middleware validasi
type PaymentProcessRequest struct {
	BookingID string `json:"booking_id" binding:"required"`
	Forced    string `json:"forced,omitempty" binding:"omitempty,oneof=SUCCESS FAILED EXPIRED"`
}

// ScannerValidateRequest untuk middleware validasi
type ScannerValidateRequest struct {
	TicketID string `json:"ticket_id" binding:"required"`
	Gate     string `json:"gate,omitempty" binding:"max=50,secure_string,no_xss"`
	Petugas  string `json:"petugas,omitempty" binding:"max=100,secure_string,no_xss"`
}

// AuthLoginRequest untuk middleware validasi
type AuthLoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=150"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// AuthRegisterRequest untuk middleware validasi
type AuthRegisterRequest struct {
	Name            string `json:"name" binding:"required,min=3,max=100,secure_string,no_xss"`
	Email           string `json:"email" binding:"required,email,max=150"`
	Phone           string `json:"phone,omitempty" binding:"omitempty,phone_id"`
	Password        string `json:"password" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// AuthForgotPasswordRequest untuk middleware validasi
type AuthForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email,max=150"`
}

// AuthResetPasswordRequest untuk middleware validasi
type AuthResetPasswordRequest struct {
	Token           string `json:"token" binding:"required"`
	Password        string `json:"password" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// AuthChangePasswordRequest untuk middleware validasi
type AuthChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=8,max=72"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// RateLimitConfig menyimpan konfigurasi rate limiting
type RateLimitConfig struct {
	Requests int           // Number of requests allowed
	Window   time.Duration // Time window
	KeyFunc  func(*gin.Context) string
}

// RateLimiterMiddleware membuat middleware rate limiting
func RateLimiterMiddleware(config RateLimitConfig) gin.HandlerFunc {
	// Ini rate limiter in-memory sederhana. Di produksi, gunakan rate limiting berbasis Redis
	return func(c *gin.Context) {
		_ = config.KeyFunc(c)
		// Implementasinya akan diletakkan di sini
		c.Next()
	}
}

// SecurityHeadersMiddleware menambahkan header keamanan
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		c.Next()
	}
}

// RequestSizeLimiterMiddleware membatasi ukuran body request
func RequestSizeLimiterMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}
