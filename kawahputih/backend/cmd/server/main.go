// Command server adalah titik masuk API. File ini hanya memuat config lalu menyerahkan ke bootstrap.go + server.go.
package main

import (
	"github.com/KawahPutih/backend/internal/config"
	applogger "github.com/KawahPutih/backend/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	applogger.Init(cfg.App.Env)
	log := applogger.L()
	defer applogger.Sync()

	db, redisClient := connectInfra(cfg, log)
	if redisClient != nil {
		defer redisClient.Close()
	}

	app := buildApp(cfg, db, log)
	runServer(cfg, app, log)
}
