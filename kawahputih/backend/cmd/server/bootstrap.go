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
	"github.com/KawahPutih/backend/internal/domain/role"
	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/internal/domain/tickettype"
	"github.com/KawahPutih/backend/internal/domain/visitorrule"
	"github.com/KawahPutih/backend/internal/seed"
	"github.com/glebarez/sqlite"
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

	// Pakai database SQLite in-memory (driver Go murni, tanpa CGO) sehingga tidak butuh DB eksternal. Data sementara dan hilang saat restart (mode demo/dev).
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		log.Fatal("failed to connect to in-memory database", zap.Error(err))
	}
	log.Info("database connected (in-memory SQLite)")

	// Auto-migrate semua entity domain, lalu seed data referensi (idempoten di DB persisten; wajib di SQLite in-memory).
	if err := migrateAndSeed(db, log); err != nil {
		log.Fatal("failed to migrate and seed database", zap.Error(err))
	}

	redisClient, err := cache.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		// Redis hanya untuk cache, bukan penentu kebenaran data inti. Bila gagal, jangan gagalkan boot, cukup catat di log.
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
		// auth + role/user
		&auth.User{}, &auth.RefreshToken{}, &auth.AuthToken{},
		&role.Role{}, &role.Permission{},
		// katalog
		&tourismpackage.Package{},
		&article.Article{},
		&gallery.Item{},
		&tickettype.TicketType{},
		// konten
		&destinationinfo.DestinationInfo{}, &faq.FAQ{}, &facility.Facility{}, &visitorrule.VisitorRule{},
		// transaksi
		&booking.Booking{}, &booking.BookingItem{},
		&payments.Payment{},
		&ticketing.Ticket{}, &ticketing.ScanRecord{},
		// audit
		&audit.AuditLog{},
	); err != nil {
		return err
	}

	// Tabel join user_roles: role.userRole tidak diekspor sehingga AutoMigrate tidak melihatnya, jadi tabelnya dibuat langsung (idempoten).
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS user_roles (
		user_id char(36) NOT NULL,
		role_id char(36) NOT NULL,
		PRIMARY KEY (user_id, role_id)
	)`).Error; err != nil {
		return err
	}

	return seed.Run(context.Background(), db, log)
}
