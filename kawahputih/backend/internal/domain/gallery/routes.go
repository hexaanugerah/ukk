package gallery

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	rg.GET("/gallery", h.List)
	rg.GET("/gallery/search", h.Search)
	rg.GET("/gallery/:id", h.Get)

	admin := rg.Group("", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	admin.POST("/gallery", h.Create)
	admin.PUT("/gallery/:id", h.Update)
	admin.DELETE("/gallery/:id", h.Delete)
}
