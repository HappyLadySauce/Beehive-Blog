package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/config"
	"github.com/zeromicro/go-zero/rest/httpx"
	"golang.org/x/time/rate"
)

type RateLimitMiddleware struct {
	enabled bool
	rps     rate.Limit
	burst   int

	mu       sync.Mutex
	visitors map[string]*visitorLimiter
}

type visitorLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimitMiddleware(conf config.RateLimitConf) *RateLimitMiddleware {
	rps := conf.RPS
	if rps <= 0 {
		rps = 50
	}
	burst := conf.Burst
	if burst <= 0 {
		burst = 100
	}
	m := &RateLimitMiddleware{
		enabled:  conf.Enabled,
		rps:      rate.Limit(rps),
		burst:    burst,
		visitors: make(map[string]*visitorLimiter),
	}
	go m.cleanupLoop()
	return m
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !m.enabled {
			next(w, r)
			return
		}

		key := clientKey(r)
		limiter := m.getLimiter(key)
		if !limiter.Allow() {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
				"code":    "RATE_LIMITED",
				"message": "too many requests",
			})
			return
		}
		next(w, r)
	}
}

func (m *RateLimitMiddleware) getLimiter(key string) *rate.Limiter {
	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()

	item, ok := m.visitors[key]
	if !ok {
		item = &visitorLimiter{
			limiter:  rate.NewLimiter(m.rps, m.burst),
			lastSeen: now,
		}
		m.visitors[key] = item
		return item.limiter
	}
	item.lastSeen = now
	return item.limiter
}

func (m *RateLimitMiddleware) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		expireBefore := time.Now().Add(-10 * time.Minute)
		m.mu.Lock()
		for key, item := range m.visitors {
			if item.lastSeen.Before(expireBefore) {
				delete(m.visitors, key)
			}
		}
		m.mu.Unlock()
	}
}

func clientKey(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		if idx := strings.IndexByte(xff, ','); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return xff
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}
