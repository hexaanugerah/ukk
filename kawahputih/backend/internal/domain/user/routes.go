package user

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)
	admin := rg.Group("/admin", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	admin.GET("/users", h.List)
	admin.PATCH("/users/:id/active", h.SetActive)
}
