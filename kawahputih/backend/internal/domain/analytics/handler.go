package analytics

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

// AdminDashboard adalah payload kartu statistik panel admin.
func (h *Handler) AdminDashboard(c *gin.Context) {
	stats, err := h.service.AdminDashboard(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "dashboard stats retrieved", stats)
}

// ManagerAnalytics adalah payload monitoring/laporan manager.
func (h *Handler) ManagerAnalytics(c *gin.Context) {
	r, err := ResolveRange(c.Query("range"), c.Query("custom"), c.Query("from"), c.Query("to"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	result, err := h.service.ManagerAnalytics(c.Request.Context(), r)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "analytics retrieved", result)
}

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	admin := rg.Group("/admin", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	admin.GET("/dashboard", h.AdminDashboard)

	manager := rg.Group("/manager", requireAuth, middleware.RequireRole("MANAGER", "ADMIN", "super_admin"))
	manager.GET("/analytics", h.ManagerAnalytics)
}
