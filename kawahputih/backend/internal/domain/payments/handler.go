package payments

import (
	"strconv"

	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase *Usecase
}

func NewHandler(usecase *Usecase) *Handler { return &Handler{usecase: usecase} }

// privilegedRead: admin dan manager boleh membaca detail pembayaran untuk laporan.
func privilegedRead(c *gin.Context) bool {
	return middleware.HasRole(c, "ADMIN") || middleware.HasRole(c, "MANAGER") || middleware.HasRole(c, "super_admin")
}

// privilegedWrite: hanya admin yang boleh memproses/mengakhiri pembayaran mana pun.
// Manager hanya memantau, pengunjung hanya boleh bertindak atas booking miliknya.
func privilegedWrite(c *gin.Context) bool {
	return middleware.HasRole(c, "ADMIN") || middleware.HasRole(c, "super_admin")
}

// Process menjalankan simulasi pembayaran QRIS untuk sebuah booking.
func (h *Handler) Process(c *gin.Context) {
	var req ProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	result, err := h.usecase.Process(c.Request.Context(), req, middleware.CurrentUserID(c), privilegedWrite(c))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment processed", result)
}

// Expire menandai pembayaran pending (dan booking-nya) kedaluwarsa.
func (h *Handler) Expire(c *gin.Context) {
	bookingID := c.Query("booking_id")
	if bookingID == "" {
		response.BadRequest(c, "booking_id query parameter is required", nil)
		return
	}
	if err := h.usecase.Expire(c.Request.Context(), bookingID, middleware.CurrentUserID(c), privilegedWrite(c)); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment expired", nil)
}

// GetByBooking mengembalikan detail gabungan payment+ticket+booking.
func (h *Handler) GetByBooking(c *gin.Context) {
	detail, err := h.usecase.GetDetail(c.Request.Context(), c.Param("bookingId"), middleware.CurrentUserID(c), privilegedRead(c))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment retrieved", detail)
}

// MarkDownloaded mencatat unduhan e-tiket untuk booking.
func (h *Handler) MarkDownloaded(c *gin.Context) {
	if err := h.usecase.MarkDownloaded(c.Request.Context(), c.Param("bookingId"), middleware.CurrentUserID(c), privilegedRead(c)); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "download recorded", nil)
}

// AdminList mendaftar semua pembayaran.
func (h *Handler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	result, total, err := h.usecase.ListAll(c.Request.Context(), page, limit, c.Query("status"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "payments retrieved", result, response.BuildMeta(page, limit, total))
}

// AdminSetStatus menerapkan override status pembayaran oleh admin.
func (h *Handler) AdminSetStatus(c *gin.Context) {
	var req SetStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	result, err := h.usecase.SetStatusByAdmin(c.Request.Context(), c.Param("id"), req.Status)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment status updated", result)
}

// GetByID mengembalikan satu pembayaran berdasarkan ID-nya.
func (h *Handler) GetByID(c *gin.Context) {
	result, err := h.usecase.GetByID(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c), privilegedRead(c))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment retrieved", result)
}

// Webhook menangani callback gateway pembayaran (Midtrans, dll). Endpoint ini publik (tanpa auth) tapi memvalidasi signature.
func (h *Handler) Webhook(c *gin.Context) {
	var req struct {
		OrderID           string `json:"order_id" form:"order_id"`
		TransactionStatus string `json:"transaction_status" form:"transaction_status"`
		FraudStatus       string `json:"fraud_status" form:"fraud_status"`
		SignatureKey      string `json:"signature_key" form:"signature_key"`
		GrossAmount       string `json:"gross_amount" form:"gross_amount"`
	}
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "invalid webhook payload", nil)
		return
	}
	if err := h.usecase.HandleMidtransWebhook(c.Request.Context(), req.OrderID, req.TransactionStatus, req.FraudStatus, req.SignatureKey, req.GrossAmount); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "webhook processed", gin.H{"status": "ok"})
}

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	payments := rg.Group("/payments", requireAuth)
	payments.POST("/process", h.Process)
	payments.POST("/expire", h.Expire)
	payments.GET("/:id", h.GetByID)
	payments.GET("/booking/:bookingId", h.GetByBooking)
	payments.POST("/booking/:bookingId/download", h.MarkDownloaded)

	// Webhook bersifat publik (tanpa auth)
	rg.POST("/payments/webhook", h.Webhook)

	admin := rg.Group("/admin", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	admin.GET("/payments", h.AdminList)
	admin.PATCH("/payments/:id/status", h.AdminSetStatus)
}
