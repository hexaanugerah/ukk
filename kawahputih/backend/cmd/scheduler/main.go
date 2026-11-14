// Command scheduler menjalankan job latar sebagai proses terpisah dari API server, berguna bila perlu di-scale sendiri. Pada skala sekarang, sweeper juga berjalan in-process di cmd/server.
package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KawahPutih/backend/internal/config"
	"github.com/KawahPutih/backend/internal/domain/booking"
	"github.com/KawahPutih/backend/internal/domain/payments"
	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/internal/scheduler"
	applogger "github.com/KawahPutih/backend/pkg/logger"
	gormlogger "gorm.io/gorm/logger"
)

// noopAuditLogger adalah audit logger no-op untuk scheduler.
type noopAuditLogger struct{}

func (noopAuditLogger) LogSuccess(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string) {
}
func (noopAuditLogger) LogFailure(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string) {
}

// noopBookingGateway adalah gateway booking no-op untuk scheduler.
type noopBookingGateway struct{}

func (noopBookingGateway) GetSnapshot(ctx context.Context, bookingID, requesterID string, privileged bool) (*booking.Snapshot, error) {
	return nil, nil
}
func (noopBookingGateway) MarkPaid(ctx context.Context, bookingID string) error { return nil }
func (noopBookingGateway) MarkExpired(ctx context.Context, bookingID string) error {
	return nil
}
func (noopBookingGateway) GetResponse(ctx context.Context, bookingID, requesterID string, privileged bool) (*booking.BookingResponse, error) {
	return nil, nil
}
func (noopBookingGateway) SetStatusByAdmin(ctx context.Context, bookingID, status string) (*booking.BookingResponse, error) {
	return nil, nil
}
func (noopBookingGateway) StalePendingIDs(ctx context.Context, cutoff time.Time, limit int) ([]string, error) {
	return nil, nil
}

// noopTicketIssuer adalah penerbit tiket no-op untuk scheduler.
type noopTicketIssuer struct{}

func (noopTicketIssuer) IssueForBooking(ctx context.Context, snap *booking.Snapshot) (*ticketing.TicketResponse, error) {
	return nil, nil
}
func (noopTicketIssuer) GetByBookingID(ctx context.Context, bookingID string) (*ticketing.TicketResponse, error) {
	return nil, nil
}

// noopMidtransGateway adalah gateway Midtrans no-op untuk scheduler.
type noopMidtransGateway struct{}

func (noopMidtransGateway) CreateTransaction(ctx context.Context, orderID string, grossAmount int64, customerName, customerEmail, itemName string) (string, error) {
	return "", nil
}
func (noopMidtransGateway) GetTransactionStatus(ctx context.Context, orderID string) (string, string, error) {
	return "", "", nil
}
func (noopMidtransGateway) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	return false
}

func (noopMidtransGateway) CreateQRIS(ctx context.Context, orderID string, grossAmount int64) (string, error) {
	return "", errors.New("midtrans is not available in scheduler")
}

func (noopMidtransGateway) Configured() bool { return false }

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	applogger.Init(cfg.App.Env)
	log := applogger.L()
	defer applogger.Sync()

	db, err := config.NewDatabase(cfg.Database, gormlogger.Warn)
	if err != nil {
		log.Fatal("scheduler: failed to connect to configured database")
	}

	paymentsRepo := payments.NewRepository(db)
	paymentsService := payments.NewService(paymentsRepo, noopBookingGateway{}, noopTicketIssuer{}, noopMidtransGateway{}, log, noopAuditLogger{})
	sweeper := scheduler.NewExpiredBookingSweeper(paymentsService, 30*time.Minute, 5*time.Minute, log)

	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		cancel()
	}()

	log.Info("scheduler starting: expired-booking sweeper")
	sweeper.Run(ctx)
	log.Info("scheduler exited gracefully")
}
