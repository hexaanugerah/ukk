package auth

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes memasang route domain ini ke grup yang diberikan. Setiap domain mengekspos fungsi ini, dan internal/routes/routes.go memanggil semuanya untuk menyusun API lengkap.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	g := rg.Group("/auth")
	g.POST("/register", middleware.RateLimit(1, 5), h.Register)
	g.POST("/login", middleware.RateLimit(1, 5), h.Login)
	g.POST("/refresh", middleware.RateLimit(2, 10), h.Refresh)
	g.POST("/logout", h.Logout)
	g.GET("/me", requireAuth, h.Me)
	g.PATCH("/profile", requireAuth, h.UpdateProfile)

	g.POST("/email/verify/request", middleware.RateLimit(1, 3), h.RequestEmailVerification)
	g.POST("/email/verify", h.VerifyEmail)

	g.POST("/password/forgot", middleware.RateLimit(1, 3), h.ForgotPassword)
	g.POST("/password/reset", h.ResetPassword)
	g.POST("/password/change", requireAuth, middleware.RateLimit(1, 5), h.ChangePassword)

	g.GET("/google/login", h.GoogleLogin)
	g.GET("/google/callback", h.GoogleCallback)
}
