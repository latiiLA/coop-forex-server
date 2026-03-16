package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(5, 10) // 5 req/sec, burst of 10

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, response.Status{
				Message: "Too many requests. Please try again later.",
				Error:   "Too many requests",
			})
			return
		}
		c.Next()
	}
}
