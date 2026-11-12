package tickettype

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)
	admin := rg.Group("/admin", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))

	rg.GET("/tickets", h.List)
	rg.GET("/tickets/active", h.ListActive)
	rg.GET("/tickets/categories", h.Categories)
	rg.GET("/tickets/code/:code", h.GetByCode)
	rg.GET("/tickets/:id", h.Get)

	admin.POST("/tickets", h.Create)
	admin.PUT("/tickets/:id", h.Update)
	admin.PATCH("/tickets/:id/status", h.SetStatus)
	admin.DELETE("/tickets/:id", h.Delete)
}
