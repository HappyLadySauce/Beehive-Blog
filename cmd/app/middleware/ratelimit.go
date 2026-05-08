package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// AuthPublicRateLimiter applies a token-bucket limit per client IP for unauthenticated auth endpoints.
// AuthPublicRateLimiter 对未认证认证类接口按客户端 IP 做令牌桶限流。
type AuthPublicRateLimiter struct {
	mu              sync.Mutex
	limiters        map[string]*clientLimiter
	r               rate.Limit
	burst           int
	staleAfter      time.Duration
	cleanupInterval time.Duration
	nextCleanup     time.Time
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewAuthPublicRateLimiter builds a limiter with the given sustained rate (events per second) and burst size.
// NewAuthPublicRateLimiter 使用给定可持续速率（事件/秒）与突发容量构造限流器。
func NewAuthPublicRateLimiter(eventsPerSecond float64, burst int) *AuthPublicRateLimiter {
	now := time.Now()
	return &AuthPublicRateLimiter{
		limiters:        make(map[string]*clientLimiter),
		r:               rate.Limit(eventsPerSecond),
		burst:           burst,
		staleAfter:      10 * time.Minute,
		cleanupInterval: time.Minute,
		nextCleanup:     now.Add(time.Minute),
	}
}

func (a *AuthPublicRateLimiter) limiterFor(ip string) *rate.Limiter {
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cleanupLocked(now)
	entry, ok := a.limiters[ip]
	if !ok {
		entry = &clientLimiter{
			limiter: rate.NewLimiter(a.r, a.burst),
		}
		a.limiters[ip] = entry
	}
	entry.lastSeen = now
	return entry.limiter
}

func (a *AuthPublicRateLimiter) cleanupLocked(now time.Time) {
	if now.Before(a.nextCleanup) {
		return
	}
	cutoff := now.Add(-a.staleAfter)
	for ip, entry := range a.limiters {
		if entry.lastSeen.Before(cutoff) {
			delete(a.limiters, ip)
		}
	}
	a.nextCleanup = now.Add(a.cleanupInterval)
}

// Len returns the current number of tracked client buckets for tests and diagnostics.
// Len 返回当前跟踪的客户端桶数量，供测试和诊断使用。
func (a *AuthPublicRateLimiter) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.limiters)
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
