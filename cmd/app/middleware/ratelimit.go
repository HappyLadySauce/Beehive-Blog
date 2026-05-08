package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// AuthPublicRateLimiter applies a token-bucket limit per client IP for unauthenticated auth endpoints.
// AuthPublicRateLimiter 对未认证认证类接口按客户端 IP 做令牌桶限流。
type AuthPublicRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	burst    int
}

// NewAuthPublicRateLimiter builds a limiter with the given sustained rate (events per second) and burst size.
// NewAuthPublicRateLimiter 使用给定可持续速率（事件/秒）与突发容量构造限流器。
func NewAuthPublicRateLimiter(eventsPerSecond float64, burst int) *AuthPublicRateLimiter {
	return &AuthPublicRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        rate.Limit(eventsPerSecond),
		burst:    burst,
	}
}

func (a *AuthPublicRateLimiter) limiterFor(ip string) *rate.Limiter {
	a.mu.Lock()
	defer a.mu.Unlock()
	lim, ok := a.limiters[ip]
	if !ok {
		lim = rate.NewLimiter(a.r, a.burst)
		a.limiters[ip] = lim
	}
	return lim
}

// GinMiddleware returns a handler that responds with HTTP 429 when the IP exceeds its budget.
// GinMiddleware 在超过预算时返回 HTTP 429。
func (a *AuthPublicRateLimiter) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.limiterFor(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "too many requests",
				"data":    nil,
			})
			return
		}
		c.Next()
	}
}
