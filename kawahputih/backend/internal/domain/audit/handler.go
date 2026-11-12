package audit

import (
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase *Usecase
}

func NewHandler(usecase *Usecase) *Handler { return &Handler{usecase: usecase} }

// List mengembalikan audit log berhalaman (khusus admin).
func (h *Handler) List(c *gin.Context) {
	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query parameters", nil)
		return
	}
	if errs := ValidateStruct(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	logs, total, err := h.usecase.List(c.Request.Context(), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "audit logs retrieved", logs, response.BuildMeta(req.Page, req.Limit, total))
}

// Get mengembalikan satu audit log berdasarkan ID (khusus admin).
func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	log, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "audit log retrieved", log)
}

// StatsByAction mengembalikan jumlah per aksi (admin/manager).
func (h *Handler) StatsByAction(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	stats, err := h.usecase.StatsByAction(c.Request.Context(), from, to)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "audit stats by action retrieved", stats)
}

// StatsByActor mengembalikan jumlah per aktor (admin/manager).
func (h *Handler) StatsByActor(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	stats, err := h.usecase.StatsByActor(c.Request.Context(), from, to)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "audit stats by actor retrieved", stats)
}

// RegisterRoutes memasang route audit. Pemanggil harus memasang middleware auth dan mount di bawah /admin/audit.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	rg.GET("", h.List)
	rg.GET("/stats/action", h.StatsByAction)
	rg.GET("/stats/actor", h.StatsByActor)
	rg.GET("/:id", h.Get)
}
