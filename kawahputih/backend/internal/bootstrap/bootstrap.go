package bootstrap

import (
	"time"

	"github.com/KawahPutih/backend/internal/config"
	"github.com/KawahPutih/backend/internal/domain/analytics"
	"github.com/KawahPutih/backend/internal/domain/article"
	"github.com/KawahPutih/backend/internal/domain/audit"
	"github.com/KawahPutih/backend/internal/domain/auth"
	"github.com/KawahPutih/backend/internal/domain/booking"
	"github.com/KawahPutih/backend/internal/domain/destinationinfo"
	"github.com/KawahPutih/backend/internal/domain/facility"
	"github.com/KawahPutih/backend/internal/domain/faq"
	"github.com/KawahPutih/backend/internal/domain/gallery"
	"github.com/KawahPutih/backend/internal/domain/package"
	"github.com/KawahPutih/backend/internal/domain/payments"
	"github.com/KawahPutih/backend/internal/domain/public"
	"github.com/KawahPutih/backend/internal/domain/role"
	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/internal/domain/tickettype"
	"github.com/KawahPutih/backend/internal/domain/user"
	"github.com/KawahPutih/backend/internal/domain/verification"
	"github.com/KawahPutih/backend/internal/domain/visitorrule"
	"github.com/KawahPutih/backend/internal/domain/visitors"

	"github.com/KawahPutih/backend/internal/payment/midtrans"
	"github.com/KawahPutih/backend/internal/scheduler"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/KawahPutih/backend/pkg/mailer"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Graph struct {
	BookingService   *booking.Service
	BookingUsecase   *booking.Usecase
	PaymentsService  *payments.Service
	TicketingService *ticketing.Service
}

type App struct {
	FrontendURL            string
	AuthHandler            *auth.Handler
	RoleHandler            *role.Handler
	UserHandler            *user.Handler
	ArticleHandler         *article.Handler
	GalleryHandler         *gallery.Handler
	PackageHandler         *tourismpackage.Handler
	BookingHandler         *booking.Handler
	TicketHandler          *tickettype.Handler
	PaymentHandler         *payments.Handler
	TicketingHandler       *ticketing.Handler
	AnalyticsHandler       *analytics.Handler
	VisitorsHandler        *visitors.Handler
	AuditHandler           *audit.Handler
	PublicHandler          *public.Handler
	VerificationHandler    *verification.Handler
	DestinationInfoHandler *destinationinfo.Handler
	FAQHandler             *faq.Handler
	FacilityHandler        *facility.Handler
	VisitorRuleHandler     *visitorrule.Handler

	JWTManager *jwtutil.Manager
	Sweeper    *scheduler.ExpiredBookingSweeper
}

func Build(cfg *config.Config, db *gorm.DB, log *zap.Logger) *App {
	jwtManager := jwtutil.NewManager(cfg.JWT.AccessSecret, cfg.JWT.RefreshSecret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL, cfg.JWT.Issuer)
	mailerClient := mailer.New(mailer.Config{Host: cfg.SMTP.Host, Port: cfg.SMTP.Port, User: cfg.SMTP.User, Password: cfg.SMTP.Password, From: cfg.SMTP.From}, log)

	_ = midtrans.NewGateway(midtrans.Config{ServerKey: cfg.Midtrans.ServerKey, ClientKey: cfg.Midtrans.ClientKey, IsProduction: cfg.Midtrans.IsProduction})

	roleRepo := role.NewRepository(db)
	roleService := role.NewService(roleRepo)
	roleUsecase := role.NewUsecase(roleService)
	roleHandler := role.NewHandler(roleUsecase)

	// domain auth
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, jwtManager, mailerClient, cfg.GoogleOAuth, cfg.App.FrontendURL, log)
	authUsecase := auth.NewUsecase(authService, roleService)
	authHandler := auth.NewHandler(authUsecase)

	// domain user (manajemen staf)
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userUsecase := user.NewUsecase(userService, roleService)
	userHandler := user.NewHandler(userUsecase)

	// domain konten
	articleRepo := article.NewRepository(db)
	articleService := article.NewService(articleRepo)
	articleHandler := article.NewHandler(article.NewUsecase(articleService))

	galleryRepo := gallery.NewRepository(db)
	galleryService := gallery.NewService(galleryRepo)
	galleryHandler := gallery.NewHandler(gallery.NewUsecase(galleryService))

	destinationInfoRepo := destinationinfo.NewRepository(db)
	destinationInfoService := destinationinfo.NewService(destinationInfoRepo)
	destinationInfoHandler := destinationinfo.NewHandler(destinationinfo.NewUsecase(destinationInfoService))

	faqRepo := faq.NewRepository(db)
	faqService := faq.NewService(faqRepo)
	faqHandler := faq.NewHandler(faq.NewUsecase(faqService))

	facilityRepo := facility.NewRepository(db)
	facilityService := facility.NewService(facilityRepo)
	facilityHandler := facility.NewHandler(facility.NewUsecase(facilityService))

	visitorRuleRepo := visitorrule.NewRepository(db)
	visitorRuleService := visitorrule.NewService(visitorRuleRepo)
	visitorRuleHandler := visitorrule.NewHandler(visitorrule.NewUsecase(visitorRuleService))

	packageRepo := tourismpackage.NewRepository(db)
	packageService := tourismpackage.NewService(packageRepo)
	packageHandler := tourismpackage.NewHandler(tourismpackage.NewUsecase(packageService))

	// domain tickettype (katalog)
	ticketTypeRepo := tickettype.NewRepository(db)
	ticketTypeService := tickettype.NewService(ticketTypeRepo)
	ticketTypeHandler := tickettype.NewHandler(ticketTypeService, nil)

	// adapter audit (dibuat lebih dulu agar service lain bisa memakainya)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)
	auditLogger := &auditLoggerAdapter{svc: auditService}

	// domain booking (multi-item)
	bookingRepo := booking.NewRepository(db)
	bookingService := new(booking.Service)
	bookingUsecase := booking.NewUsecase(bookingService, &ticketPricerAdapter{svc: ticketTypeService, sold: bookingRepo})
	bookingHandler := booking.NewHandler(bookingUsecase, nil)

	// domain ticketing (e-tiket + scanner)
	ticketingRepo := ticketing.NewRepository(db)
	ticketingService := ticketing.NewService(ticketingRepo, &bookingCheckinAdapter{uc: bookingUsecase}, log, auditLogger)
	ticketingUsecase := ticketing.NewUsecase(ticketingService)
	ticketingHandler := ticketing.NewHandler(ticketingUsecase)

	// domain payments
	paymentsRepo := payments.NewRepository(db)
	midtransGateway := midtrans.NewGateway(midtrans.Config{ServerKey: cfg.Midtrans.ServerKey, ClientKey: cfg.Midtrans.ClientKey, IsProduction: cfg.Midtrans.IsProduction})
	paymentsService := payments.NewService(paymentsRepo, &bookingGatewayAdapter{uc: bookingUsecase}, &ticketIssuerAdapter{svc: ticketingService}, midtransGateway, log, auditLogger)
	paymentsUsecase := payments.NewUsecase(paymentsService)
	paymentsHandler := payments.NewHandler(paymentsUsecase)

	// isi bagian yang terikat belakangan (booking butuh payments + ticketing)
	*bookingService = *booking.NewService(bookingRepo,
		&paymentStarterAdapter{svc: paymentsService},
		&ticketStatusSyncerAdapter{svc: ticketingService},
		auditLogger,
	)
	bookingHandler = booking.NewHandler(bookingUsecase, bookingPaymentReader(paymentsService))

	ticketTypeHandler = tickettype.NewHandler(ticketTypeService, &bookingSalesAdapter{lag: bookingRepo})

	// domain analytics
	analyticsRepo := analytics.NewRepository(db)
	analyticsService := analytics.NewService(analyticsRepo)
	analyticsHandler := analytics.NewHandler(analyticsService)

	// domain audit
	auditUsecase := audit.NewUsecase(auditService)
	auditHandler := audit.NewHandler(auditUsecase)

	// domain public
	publicRepo := public.NewRepository(db)
	publicService := public.NewService(publicRepo, ticketTypeService)
	publicHandler := public.NewHandler(publicService)

	// domain verification
	verificationRepo := verification.NewRepository(db)
	verificationService := verification.NewService(verificationRepo, ticketingService)
	verificationHandler := verification.NewHandler(verificationService)

	// domain visitors/staff
	visitorsRepo := visitors.NewRepository(db)
	visitorsService := visitors.NewService(visitorsRepo)
	visitorsHandler := visitors.NewHandler(visitorsService)

	sweeper := scheduler.NewExpiredBookingSweeper(paymentsService, 30*time.Minute, 5*time.Minute, log)

	return &App{
		FrontendURL: cfg.App.FrontendURL,
		AuthHandler: authHandler, RoleHandler: roleHandler, UserHandler: userHandler,
		ArticleHandler: articleHandler, GalleryHandler: galleryHandler, PackageHandler: packageHandler,
		BookingHandler: bookingHandler, TicketHandler: ticketTypeHandler, PaymentHandler: paymentsHandler,
		TicketingHandler: ticketingHandler, AnalyticsHandler: analyticsHandler, VisitorsHandler: visitorsHandler,
		AuditHandler: auditHandler, PublicHandler: publicHandler, VerificationHandler: verificationHandler,
		DestinationInfoHandler: destinationInfoHandler, FAQHandler: faqHandler, FacilityHandler: facilityHandler,
		VisitorRuleHandler: visitorRuleHandler,
		JWTManager:         jwtManager, Sweeper: sweeper,
	}
}
