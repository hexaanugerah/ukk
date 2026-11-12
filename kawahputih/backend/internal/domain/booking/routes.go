package booking

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	// Pengunjung: membuat, membatalkan, dan melihat booking sendiri. Booking hanya untuk pengguna yang sudah masuk (tamu harus daftar dulu).
	bookings := rg.Group("/bookings", requireAuth)
	bookings.POST("", h.Create)
	bookings.GET("/mine", h.ListMine)
	bookings.GET("/:id", h.Get)
	bookings.GET("/code/:code", h.GetByCode)
	bookings.POST("/:id/cancel", h.Cancel)
	bookings.DELETE("/:id", h.Cancel) // backward compatibility

	// Admin + Manager: pelaporan booking hanya-baca. Manager menjalankan monitoring dan laporan pendapatan, jadi boleh mendaftar booking tapi tidak mengubahnya.
	reports := rg.Group("/admin", requireAuth, middleware.RequireRole("ADMIN", "MANAGER", "super_admin"))
	reports.GET("/bookings", h.ListAll)
	reports.GET("/bookings/today", h.ListToday)
	reports.GET("/visitors/:id/bookings", h.ListForVisitor)

	// Khusus admin: perubahan siklus hidup booking.
	admin := rg.Group("/admin", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	admin.PATCH("/bookings/:id/status", h.SetStatus)
}
