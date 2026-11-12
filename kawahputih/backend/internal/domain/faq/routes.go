package faq

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	// Endpoint publik
	public := rg.Group("/faq")
	public.GET("", h.ListPublished)
	public.GET("/:id", h.Get)

	// Endpoint admin
	requireAuth := middleware.RequireAuth(jwtManager)
	content := rg.Group("/admin/faq", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	content.POST("", h.Create)
	content.PUT("/:id", h.Update)
	content.DELETE("/:id", h.Delete)
	content.GET("", h.List)
	content.GET("/:id", h.Get)
}
