package main

import (
	"context"

	"github.com/KawahPutih/backend/internal/bootstrap"
	"github.com/KawahPutih/backend/internal/cache"
	"github.com/KawahPutih/backend/internal/config"
	"github.com/KawahPutih/backend/internal/domain/article"
	"github.com/KawahPutih/backend/internal/domain/audit"
	"github.com/KawahPutih/backend/internal/domain/auth"
	"github.com/KawahPutih/backend/internal/domain/booking"
	"github.com/KawahPutih/backend/internal/domain/destinationinfo"
	"github.com/KawahPutih/backend/internal/domain/facility"
	"github.com/KawahPutih/backend/internal/domain/faq"
	"github.com/KawahPutih/backend/internal/domain/gallery"
	tourismpackage "github.com/KawahPutih/backend/internal/domain/package"
	"github.com/KawahPutih/backend/internal/domain/payments"
	publicdomain "github.com/KawahPutih/backend/internal/domain/public"
	"github.com/KawahPutih/backend/internal/domain/role"
	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/internal/domain/tickettype"
	"github.com/KawahPutih/backend/internal/domain/visitorrule"
	"github.com/KawahPutih/backend/internal/seed"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func connectInfra(cfg *config.Config, log *zap.Logger) (*gorm.DB, *redis.Client) {
	gormLogLevel := gormlogger.Warn
	if cfg.App.Env == "local" {
		gormLogLevel = gormlogger.Info
	}
	db, err := config.NewDatabase(cfg.Database, gormLogLevel)
	if err != nil {
		log.Fatal("failed to connect to configured database")
	}
	log.Info("database connected")

	if err := migrateAndSeed(db, log); err != nil {
		log.Fatal("failed to migrate and seed database", zap.Error(err))
	}

	redisClient, err := cache.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Warn("redis unavailable, continuing without cache", zap.Error(err))
		return db, nil
	}
	log.Info("redis connected", zap.String("addr", cfg.Redis.Addr))
	return db, redisClient
}

func buildApp(cfg *config.Config, db *gorm.DB, log *zap.Logger) *bootstrap.App {
	return bootstrap.Build(cfg, db, log)
}

func migrateAndSeed(db *gorm.DB, log *zap.Logger) error {
	if err := db.AutoMigrate(
		&auth.User{}, &auth.RefreshToken{}, &auth.AuthToken{},
		&role.Role{}, &role.Permission{},
		&tourismpackage.Package{},
		&article.Article{},
		&gallery.Item{},
		&tickettype.TicketType{},
		&destinationinfo.DestinationInfo{}, &faq.FAQ{}, &facility.Facility{}, &visitorrule.VisitorRule{},
		&publicdomain.Destination{}, &publicdomain.Review{}, &publicdomain.Content{},
		&booking.Booking{}, &booking.BookingItem{},
		&payments.Payment{}, &payments.PaymentDestination{},
		&payments.PaymentSetting{}, &payments.PaymentMethodConfig{}, &payments.PaymentWebhookLog{},
		&ticketing.Ticket{}, &ticketing.ScanRecord{},
		&audit.AuditLog{},
	); err != nil {
		return err
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS user_roles (
		user_id char(36) NOT NULL,
		role_id char(36) NOT NULL,
		PRIMARY KEY (user_id, role_id)
	)`).Error; err != nil {
		return err
	}
	return seed.Run(context.Background(), db, log)
}
