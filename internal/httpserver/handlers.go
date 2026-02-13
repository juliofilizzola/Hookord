package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func HealthCheck(logger *zerolog.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		logger.Debug().
			Str("Request_id", context.GetString("request_id")).
			Msg("Health check endpoint called")

		context.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func Metrics(logger *zerolog.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		logger.Debug().
			Str("Request_id", context.GetString("request_id")).
			Msg("Metrics endpoint called")

		context.JSON(http.StatusOK, gin.H{
			"uptime":           "24h",
			"events_processed": 1000,
			"version":          "0.0.1",
		})
	}
}
