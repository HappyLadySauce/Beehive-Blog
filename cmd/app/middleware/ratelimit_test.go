package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAuthPublicRateLimiterRejectsWhenBudgetExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewAuthPublicRateLimiter(0.1, 1)
	router := gin.New()
	router.GET("/", limiter.GinMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	first := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	router.ServeHTTP(first, req)
	if first.Code != http.StatusNoContent {
		t.Fatalf("first status = %d, want %d", first.Code, http.StatusNoContent)
	}

	second := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	router.ServeHTTP(second, req)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d", second.Code, http.StatusTooManyRequests)
	}
}

func TestAuthPublicRateLimiterCleansIdleBuckets(t *testing.T) {
	limiter := NewAuthPublicRateLimiter(1, 1)
	_ = limiter.limiterFor("203.0.113.10")
	if got := limiter.Len(); got != 1 {
		t.Fatalf("tracked limiters = %d, want 1", got)
	}

	limiter.mu.Lock()
	limiter.limiters["203.0.113.10"].lastSeen = time.Now().Add(-11 * time.Minute)
	limiter.nextCleanup = time.Now().Add(-time.Second)
	limiter.mu.Unlock()

	_ = limiter.limiterFor("203.0.113.11")
	if got := limiter.Len(); got != 1 {
		t.Fatalf("tracked limiters after cleanup = %d, want 1", got)
	}
}
