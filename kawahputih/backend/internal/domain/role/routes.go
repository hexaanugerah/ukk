package role

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes memasang manajemen role di bawah /super-admin; hanya super_admin yang mengelola penetapan role.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	superAdmin := rg.Group("/super-admin", requireAuth, middleware.RequireRole("super_admin"))
	superAdmin.GET("/roles", h.List)
	superAdmin.PUT("/users/:userId/roles", h.AssignToUser)

	owner := rg.Group("/owner", requireAuth, middleware.RequireRole("owner", "super_admin"))
	owner.GET("/roles", h.List) // read-only for Owner
}
