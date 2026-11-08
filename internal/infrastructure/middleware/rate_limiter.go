package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"golang.org/x/time/rate"
)

// Per-client SPA traffic needs headroom: pages often fire several GETs at once,
// and browsers also send CORS preflights. A single global 5 rps limiter will 429
// normal dashboard loads.
const (
	rateLimitRPS   = 30
	rateLimitBurst = 60
	limiterTTL     = 10 * time.Minute
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func init() {
	go cleanupVisitors()
}

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(rateLimitRPS), rateLimitBurst)
		visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > limiterTTL {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Preflight requests are not real API usage; don't burn the budget.
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		ip := c.ClientIP()
		if !getVisitor(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, response.Status{
				Message: "Too many requests. Please try again later.",
				Error:   "Too many requests",
			})
			return
		}
		c.Next()
	}
}
