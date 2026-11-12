package visitors

import (
	"strconv"

	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

// pengunjung

func (h *Handler) ListVisitors(c *gin.Context) {
	rows, err := h.service.repo.ListVisitors(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	response.OK(c, "visitors retrieved", rows)
}

func (h *Handler) GetVisitor(c *gin.Context) {
	row, err := h.service.repo.FindVisitorByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	if row == nil {
		response.NotFound(c, "visitor not found")
		return
	}
	response.OK(c, "visitor retrieved", row)
}

func (h *Handler) VisitorStats(c *gin.Context) {
	stats, err := h.service.repo.VisitorStats(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	response.OK(c, "visitor stats retrieved", stats)
}

// staf

func (h *Handler) ListStaff(c *gin.Context) {
	rows, err := h.service.repo.ListStaff(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	response.OK(c, "staff retrieved", rows)
}

func (h *Handler) GetStaff(c *gin.Context) {
	row, err := h.service.repo.FindStaffByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	if row == nil {
		response.NotFound(c, "staff not found")
		return
	}
	response.OK(c, "staff retrieved", row)
}

func (h *Handler) StaffStats(c *gin.Context) {
	stats, err := h.service.repo.StaffStats(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	response.OK(c, "staff stats retrieved", stats)
}

func (h *Handler) SetStaffStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if req.Status != "AKTIF" && req.Status != "NONAKTIF" {
		response.ValidationFailed(c, []response.FieldError{{Field: "status", Message: "status must be AKTIF or NONAKTIF"}})
		return
	}
	if err := h.service.repo.SetStaffActive(c.Request.Context(), c.Param("id"), req.Status == "AKTIF"); err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	response.OK(c, "staff status updated", nil)
}

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return page, limit
}

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	// Data hanya-baca yang dipakai admin DAN manager: manager menjalankan monitoring dan pelaporan, jadi harus bisa mendaftar staf/pengunjung. Mutasi tetap khusus admin.
	read := rg.Group("/admin", requireAuth, middleware.RequireRole("ADMIN", "MANAGER", "super_admin"))
	read.GET("/visitors", h.ListVisitors)
	read.GET("/visitors/stats", h.VisitorStats)
	read.GET("/visitors/:id", h.GetVisitor)
	read.GET("/staff", h.ListStaff)
	read.GET("/staff/stats", h.StaffStats)
	read.GET("/staff/:id", h.GetStaff)

	admin := rg.Group("/admin", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	admin.PATCH("/staff/:id/status", h.SetStaffStatus)
}
