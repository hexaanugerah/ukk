// Package scheduler menjalankan job latar berkala. Sweeper booking kedaluwarsa adalah satu-satunya yang terpasang di biner server; juga tersedia sebagai loop mandiri.
package scheduler

import (
	"context"
	"time"

	"github.com/KawahPutih/backend/internal/domain/payments"
	"go.uber.org/zap"
)

// ExpiredBookingSweeper secara berkala mengakhiri booking/pembayaran yang masih pending melewati jendela pembayaran. Service payments yang memiliki transisi; sweeper hanya memutuskan KAPAN.
type ExpiredBookingSweeper struct {
	payments *payments.Service
	window   time.Duration
	interval time.Duration
	log      *zap.Logger
}

func NewExpiredBookingSweeper(payments *payments.Service, window, interval time.Duration, log *zap.Logger) *ExpiredBookingSweeper {
	return &ExpiredBookingSweeper{payments: payments, window: window, interval: interval, log: log}
}

// Run memblokir dan berdetak selamanya; panggil di goroutine sendiri.
func (s *ExpiredBookingSweeper) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sweepOnce(ctx)
		}
	}
}

func (s *ExpiredBookingSweeper) sweepOnce(ctx context.Context) {
	cutoff := time.Now().Add(-s.window)
	swept, err := s.payments.ExpireStalePending(ctx, cutoff, 200)
	if err != nil {
		s.log.Error("expired booking sweep failed", zap.Error(err))
		return
	}
	if swept > 0 {
		s.log.Info("expired-booking sweep", zap.Int("swept", swept))
	}
}
