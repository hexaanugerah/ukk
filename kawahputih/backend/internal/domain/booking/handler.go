package booking

import (
	"strconv"
	"time"

	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase       *Usecase
	paymentReader PaymentReader
}

func NewHandler(usecase *Usecase, paymentReader PaymentReader) *Handler {
	return &Handler{usecase: usecase, paymentReader: paymentReader}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	result, err := h.usecase.CreateBooking(c.Request.Context(), middleware.CurrentUserID(c), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.Created(c, "booking created, proceed to payment", result)
}

func (h *Handler) Get(c *gin.Context) {
	privileged := middleware.HasRole(c, "ADMIN") || middleware.HasRole(c, "MANAGER")
	result, err := h.usecase.Get(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c), privileged, h.paymentReader)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "booking retrieved", result)
}

func (h *Handler) GetByCode(c *gin.Context) {
	privileged := middleware.HasRole(c, "ADMIN") || middleware.HasRole(c, "MANAGER")
	result, err := h.usecase.GetByCode(c.Request.Context(), c.Param("code"), middleware.CurrentUserID(c), privileged, h.paymentReader)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "booking retrieved", result)
}

func (h *Handler) Cancel(c *gin.Context) {
	privileged := middleware.HasRole(c, "ADMIN") || middleware.HasRole(c, "MANAGER")
	if err := h.usecase.Cancel(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c), privileged); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "booking cancelled", nil)
}

func (h *Handler) ListMine(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	result, total, err := h.usecase.ListMine(c.Request.Context(), middleware.CurrentUserID(c), page, limit, c.Query("status"), h.paymentReader)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "bookings retrieved", result, response.BuildMeta(page, limit, total))
}

func (h *Handler) ListAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	result, total, err := h.usecase.ListAll(c.Request.Context(), page, limit, c.Query("status"), c.Query("search"), h.paymentReader)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "bookings retrieved", result, response.BuildMeta(page, limit, total))
}

// ListForVisitor mengembalikan riwayat booking + pembayaran milik seorang pengunjung (admin/manager).
func (h *Handler) ListForVisitor(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	result, total, err := h.usecase.ListForUser(c.Request.Context(), c.Param("id"), page, limit, c.Query("status"), h.paymentReader)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "visitor bookings retrieved", result, response.BuildMeta(page, limit, total))
}

func (h *Handler) ListToday(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	result, total, err := h.usecase.ListToday(c.Request.Context(), date, page, limit, h.paymentReader)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "bookings retrieved", result, response.BuildMeta(page, limit, total))
}

func (h *Handler) SetStatus(c *gin.Context) {
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	result, err := h.usecase.SetStatus(c.Request.Context(), c.Param("id"), req.Status)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "booking status updated", result)
}
