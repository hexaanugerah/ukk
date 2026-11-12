package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/KawahPutih/backend/internal/bootstrap"
	"github.com/KawahPutih/backend/internal/config"
	"go.uber.org/zap"
)

func runServer(cfg *config.Config, app *bootstrap.App, log *zap.Logger) {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := buildRouter(app, log)

	srv := &http.Server{
		Addr: ":" + cfg.App.Port, Handler: router,
		ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
	}

	// Sweeper booking kedaluwarsa berjalan in-process di API server untuk skala sekarang (Docker satu biner). Biner cmd/scheduler tersedia bila perlu dipisah.
	schedulerCtx, cancelScheduler := context.WithCancel(context.Background())
	go app.Sweeper.Run(schedulerCtx)

	go func() {
		log.Info("server starting", zap.String("port", cfg.App.Port), zap.String("env", cfg.App.Env))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed to start", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	cancelScheduler()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("forced shutdown", zap.Error(err))
	}
	log.Info("server exited gracefully")
}

