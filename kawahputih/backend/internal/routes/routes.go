// Package routes assembles the full API by calling every domain's own
// File ini sengaja satu-satunya tempat yang mengimpor semua package domain sekaligus; domain tidak saling mengimpor, tapi harus ada yang tahu semuanya ada.
package routes

import (
	"net/http"

	"github.com/KawahPutih/backend/internal/bootstrap"
	"github.com/KawahPutih/backend/internal/domain/analytics"
	"github.com/KawahPutih/backend/internal/domain/article"
	"github.com/KawahPutih/backend/internal/domain/audit"
	"github.com/KawahPutih/backend/internal/domain/auth"
	"github.com/KawahPutih/backend/internal/domain/booking"
	"github.com/KawahPutih/backend/internal/domain/destinationinfo"
	"github.com/KawahPutih/backend/internal/domain/facility"
	"github.com/KawahPutih/backend/internal/domain/faq"
	"github.com/KawahPutih/backend/internal/domain/gallery"
	ttourismpackage "github.com/KawahPutih/backend/internal/domain/package"
	"github.com/KawahPutih/backend/internal/domain/payments"
	"github.com/KawahPutih/backend/internal/domain/public"
	"github.com/KawahPutih/backend/internal/domain/role"
	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/internal/domain/tickettype"
	"github.com/KawahPutih/backend/internal/domain/user"
	"github.com/KawahPutih/backend/internal/domain/verification"
	"github.com/KawahPutih/backend/internal/domain/visitorrule"
	"github.com/KawahPutih/backend/internal/domain/visitors"
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter(app *bootstrap.App, log *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(middleware.CORS(app.FrontendURL))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.RequestID())
	r.Use(middleware.TieredRateLimit(app.JWTManager))
	r.Use(middleware.AuditMiddleware(app.JWTManager))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	// Berkas unggahan galeri (dibuat oleh admin) disajikan dari disk.
	r.Static("/uploads", app.UploadDir)

	v1 := r.Group("/api/v1")

	auth.RegisterRoutes(v1, app.AuthHandler, app.JWTManager)
	role.RegisterRoutes(v1, app.RoleHandler, app.JWTManager)
	user.RegisterRoutes(v1, app.UserHandler, app.JWTManager)
	article.RegisterRoutes(v1, app.ArticleHandler, app.JWTManager)
	gallery.RegisterRoutes(v1, app.GalleryHandler, app.JWTManager)
	ttourismpackage.RegisterRoutes(v1, app.PackageHandler, app.JWTManager)
	booking.RegisterRoutes(v1, app.BookingHandler, app.JWTManager)
	tickettype.RegisterRoutes(v1, app.TicketHandler, app.JWTManager)
	payments.RegisterRoutes(v1, app.PaymentHandler, app.PaymentDestinationHandler, app.PaymentSettingsHandler, app.UploadDir, app.JWTManager)
	ticketing.RegisterRoutes(v1, app.TicketingHandler, app.JWTManager)
	analytics.RegisterRoutes(v1, app.AnalyticsHandler, app.JWTManager)
	visitors.RegisterRoutes(v1, app.VisitorsHandler, app.JWTManager)
	public.RegisterRoutes(v1, app.PublicHandler)
	verification.RegisterRoutes(v1, app.VerificationHandler, app.JWTManager)
	destinationinfo.RegisterRoutes(v1, app.DestinationInfoHandler, app.JWTManager)
	faq.RegisterRoutes(v1, app.FAQHandler, app.JWTManager)
	facility.RegisterRoutes(v1, app.FacilityHandler, app.JWTManager)
	visitorrule.RegisterRoutes(v1, app.VisitorRuleHandler, app.JWTManager)

	requireAuth := middleware.RequireAuth(app.JWTManager)
	requireAdmin := middleware.RequireRole("ADMIN", "super_admin")
	auditGroup := v1.Group("/admin/audit", requireAuth, requireAdmin)
	audit.RegisterRoutes(auditGroup, app.AuditHandler)

	return r
}
