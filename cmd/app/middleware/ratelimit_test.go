package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
)

func TestAuthPublicRateLimiterRejectsWhenBudgetExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := middleware.NewAuthPublicRateLimiter(0.1, 1)
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

func TestAuthPublicRateLimiterTracksClientBuckets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := middleware.NewAuthPublicRateLimiter(10, 10)
	router := gin.New()
	router.GET("/", limiter.GinMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for _, ip := range []string{"203.0.113.10:12345", "203.0.113.11:12345"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status for %s = %d, want %d", ip, rec.Code, http.StatusNoContent)
		}
	}
	if got := limiter.Len(); got != 2 {
		t.Fatalf("tracked limiters = %d, want 2", got)
	}
}
