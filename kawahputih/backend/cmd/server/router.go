package main

import (
	"github.com/gin-gonic/gin"
	"github.com/KawahPutih/backend/internal/bootstrap"
	"github.com/KawahPutih/backend/internal/routes"
	"go.uber.org/zap"
)

func buildRouter(app *bootstrap.App, log *zap.Logger) *gin.Engine {
	return routes.NewRouter(app, log)
}

