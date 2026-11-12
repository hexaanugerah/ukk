package tourismpackage

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	rg.GET("/packages", h.List)
	rg.GET("/packages/:id", h.Get)

	requireAuth := middleware.RequireAuth(jwtManager)
	admin := rg.Group("/packages", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	admin.POST("", h.Create)
	admin.PUT("/:id", h.Update)
	admin.DELETE("/:id", h.Delete)
}
