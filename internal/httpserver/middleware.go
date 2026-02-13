package httpserver

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func LoggingMiddleware(logger *zerolog.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		reqId := context.GetHeader("X-Request-Id")
		if reqId == "" {
			reqId = uuid.NewString()
		}

		context.Set("request_id", reqId)

		start := time.Now()

		path := context.Request.URL.Path

		context.Next()

		latency := time.Since(start)

		logger.Info().
			Str("request_id", reqId).
			Str("method", context.Request.Method).
			Str("path", path).
			Int("status", context.Writer.Status()).
			Dur("latency", latency).
			Int("size", context.Writer.Size()).
			Msg("HTTP Request completed")
	}
}

func RecoveryMiddleware(logger *zerolog.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error().
					Interface("panic", err).
					Str("request_id", context.GetString("request_id")).
					Str("method", context.Request.Method).
					Str("path", context.Request.URL.Path).
					Msg("Panic recovered in HTTP handler")
				context.AbortWithStatus(500)
				context.JSON(500, gin.H{"error": "Internal server error"})
			}
		}()

		context.Next()
	}
}
