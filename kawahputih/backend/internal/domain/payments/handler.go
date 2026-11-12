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

func privilegedRead(c *gin.Context) bool {
	return middleware.HasRole(c, "ADMIN") || middleware.HasRole(c, "MANAGER") || middleware.HasRole(c, "super_admin")
}

func privilegedWrite(c *gin.Context) bool {
	return middleware.HasRole(c, "ADMIN") || middleware.HasRole(c, "super_admin")
}

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
	if req.Method != "" && !ValidMethod(req.Method) {
		response.ValidationFailed(c, []response.FieldError{{Field: "method", Message: "Metode pembayaran tidak didukung."}})
		return
	}
	result, err := h.usecase.Process(c.Request.Context(), req, middleware.CurrentUserID(c), privilegedWrite(c))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment processed", result)
}

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

func (h *Handler) GetByBooking(c *gin.Context) {
	detail, err := h.usecase.GetDetail(c.Request.Context(), c.Param("bookingId"), middleware.CurrentUserID(c), privilegedRead(c))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment retrieved", detail)
}

func (h *Handler) MarkDownloaded(c *gin.Context) {
	if err := h.usecase.MarkDownloaded(c.Request.Context(), c.Param("bookingId"), middleware.CurrentUserID(c), privilegedRead(c)); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "download recorded", nil)
}

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

func (h *Handler) GetByID(c *gin.Context) {
	result, err := h.usecase.GetByID(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c), privilegedRead(c))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment retrieved", result)
}

func (h *Handler) Webhook(c *gin.Context) {
	var req struct {
		OrderID           string `json:"order_id" form:"order_id"`
		TransactionStatus string `json:"transaction_status" form:"transaction_status"`
		StatusCode        string `json:"status_code" form:"status_code"`
		FraudStatus       string `json:"fraud_status" form:"fraud_status"`
		SignatureKey      string `json:"signature_key" form:"signature_key"`
		GrossAmount       string `json:"gross_amount" form:"gross_amount"`
	}
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "invalid webhook payload", nil)
		return
	}
	if err := h.usecase.HandleMidtransWebhook(c.Request.Context(), req.OrderID, req.TransactionStatus, req.StatusCode, req.FraudStatus, req.SignatureKey, req.GrossAmount); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "webhook processed", gin.H{"status": "ok"})
}

func (h *Handler) AdminSetEWalletNumber(c *gin.Context) {
	var req SetEWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	if !ValidEWalletNumber(req.EWalletNumber) {
		response.ValidationFailed(c, []response.FieldError{{Field: "ewallet_number", Message: "Nomor e-Wallet tidak dikenal."}})
		return
	}
	result, err := h.usecase.SetEWalletNumber(c.Request.Context(), c.Param("id"), req.EWalletNumber)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "nomor e-Wallet diperbarui", result)
}

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	payments := rg.Group("/payments", requireAuth)
	payments.POST("/process", h.Process)
	payments.POST("/expire", h.Expire)
	payments.GET("/:id", h.GetByID)
	payments.GET("/booking/:bookingId", h.GetByBooking)
	payments.POST("/booking/:bookingId/download", h.MarkDownloaded)

	rg.POST("/payments/webhook", h.Webhook)

	admin := rg.Group("/admin", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	admin.GET("/payments", h.AdminList)
	admin.PATCH("/payments/:id/status", h.AdminSetStatus)
	admin.PATCH("/payments/:id/ewallet-number", h.AdminSetEWalletNumber)
}
