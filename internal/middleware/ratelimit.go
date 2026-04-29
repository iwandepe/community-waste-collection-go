package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// PickupRateLimiter limits pickup creation per IP using a token bucket.
// Each IP gets 5 requests/second with a burst of 10.
type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

func newIPLimiter(r rate.Limit, b int) *ipLimiter {
	return &ipLimiter{limiters: make(map[string]*rate.Limiter), r: r, b: b}
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.limiters[ip]; ok {
		return lim
	}
	lim := rate.NewLimiter(l.r, l.b)
	l.limiters[ip] = lim
	return lim
}

var pickupLimiter = newIPLimiter(5, 10)

func PickupRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !pickupLimiter.get(c.ClientIP()).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "too many requests, slow down",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
