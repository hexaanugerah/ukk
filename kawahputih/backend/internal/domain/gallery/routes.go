package gallery

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	// Endpoint publik
	rg.GET("/gallery", h.List)
	rg.GET("/gallery/:id", h.Get)

	// Endpoint admin
	requireAuth := middleware.RequireAuth(jwtManager)
	content := rg.Group("", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	content.POST("/gallery", h.Create)
	content.PUT("/gallery/:id", h.Update)
	content.DELETE("/gallery/:id", h.Delete)
}
