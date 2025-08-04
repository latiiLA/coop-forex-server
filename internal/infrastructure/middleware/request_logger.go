package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // for UUID generation

	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	log "github.com/sirupsen/logrus"
)

const TraceIDKey = "TraceID"

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Generate a new Trace ID or check if it exists in headers
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		// Store TraceID in context for downstream handlers
		c.Set(TraceIDKey, traceID)

		// Set it in response header so clients can see it
		c.Writer.Header().Set("X-Trace-ID", traceID)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		userID, err := utils.GetUserID(c)
		var userIDStr string
		if err != nil {
			userIDStr = "anonymous"
		} else {
			userIDStr = userID.Hex()
		}

		entry := log.WithFields(log.Fields{
			"trace_id":   traceID,
			"user_id":    userIDStr,
			"method":     c.Request.Method,
			"path":       path,
			"status":     status,
			"latency":    latency,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		})

		switch {
		case status >= 500:
			entry.Error("Server error")
		case status >= 400:
			entry.Warn("Client error")
		default:
			entry.Info("Request completed")
		}
	}
}
