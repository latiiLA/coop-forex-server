package utils

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetLogger returns a logrus Entry with common request fields from the context
func GetLogger(c *gin.Context) *log.Entry {
	traceID := c.GetString("TraceID")
	if traceID == "" {
		traceID = "unknown"
	}

	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	userID := "anonymous"
	if uid, err := GetUserID(c); err == nil {
		userID = uid.Hex()
	}

	return log.WithFields(log.Fields{
		"trace_id":   traceID,
		"user_id":    userID,
		"ip":         clientIP,
		"user_agent": userAgent,
	})
}
