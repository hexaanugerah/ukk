package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func Init(env string) {
	var cfg zap.Config
	if env == "local" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	l, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("logger: failed to initialize zap logger: " + err.Error())
	}
	log = l
}

func L() *zap.Logger {
	if log == nil {
		log, _ = zap.NewProduction()
	}
	return log
}

func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}
