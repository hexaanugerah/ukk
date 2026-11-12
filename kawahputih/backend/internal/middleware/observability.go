package middleware

import (
	"time"

	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("request",
			zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()), zap.Duration("latency", time.Since(start)), zap.String("ip", c.ClientIP()))
	}
}

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered", zap.Any("error", r), zap.String("path", c.Request.URL.Path))
				response.InternalError(c, "")
				c.Abort()
			}
		}()
		c.Next()
	}
}
