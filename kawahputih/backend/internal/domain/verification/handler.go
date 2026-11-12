package verification

import (
	"strconv"

	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Verify(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	staffID := middleware.CurrentUserID(c)
	staffName := req.Petugas
	if staffName == "" {
		staffName = staffID
	}
	result, err := h.service.Verify(c.Request.Context(), req, staffID, staffName)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, result.Message, result)
}

func (h *Handler) GetVerification(c *gin.Context) {
	result, err := h.service.GetVerification(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "verification retrieved", result)
}

func (h *Handler) ListVerifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	gate := c.Query("gate")
	petugas := c.Query("petugas")
	from := c.Query("from")
	to := c.Query("to")
	items, total, err := h.service.ListVerifications(c.Request.Context(), page, limit, gate, petugas, from, to)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "verifications retrieved", items, response.BuildMeta(page, limit, total))
}

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	// Endpoint verifikasi lama hanya untuk petugas, sama seperti scanner di domain
	// ticketing: admin/manager/pengunjung tidak boleh memvalidasi atau membaca
	// riwayat pemindaian mentah.
	petugas := middleware.RequireRole("PETUGAS")
	verifications := rg.Group("/verifications", requireAuth, petugas)
	verifications.POST("", h.Verify)
	verifications.GET("", h.ListVerifications)
	verifications.GET("/:id", h.GetVerification)

	// Alias lama untuk memindai tiket.
	tickets := rg.Group("/tickets", requireAuth, petugas)
	tickets.POST("/verify", h.Verify)
}
