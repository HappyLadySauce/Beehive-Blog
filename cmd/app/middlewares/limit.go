package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

const (
	defaultGlobalLimitPerMinute = 120
	defaultGlobalWindow         = time.Minute
	defaultLoginFailLimit       = 8
	defaultLoginFailWindow      = 15 * time.Minute
)

type windowCounter struct {
	count      int
	windowEnds time.Time
}

type localWindowLimiter struct {
	mu      sync.Mutex
	counter map[string]windowCounter
	limit   int
	window  time.Duration
}

func newLocalWindowLimiter(limit int, window time.Duration) *localWindowLimiter {
	if limit <= 0 {
		limit = 1
	}
	if window <= 0 {
		window = time.Second
	}
	return &localWindowLimiter{
		counter: make(map[string]windowCounter, 1024),
		limit:   limit,
		window:  window,
	}
}

func (l *localWindowLimiter) Allow(key string, now time.Time) bool {
	if strings.TrimSpace(key) == "" {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	item, exists := l.counter[key]
	if !exists || now.After(item.windowEnds) {
		l.counter[key] = windowCounter{
			count:      1,
			windowEnds: now.Add(l.window),
		}
		return true
	}
	if item.count >= l.limit {
		return false
	}
	item.count++
	l.counter[key] = item
	return true
}

func (l *localWindowLimiter) Increment(key string, now time.Time) int {
	if strings.TrimSpace(key) == "" {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	item, exists := l.counter[key]
	if !exists || now.After(item.windowEnds) {
		item = windowCounter{
			count:      1,
			windowEnds: now.Add(l.window),
		}
		l.counter[key] = item
		return item.count
	}

	item.count++
	l.counter[key] = item
	return item.count
}

func (l *localWindowLimiter) Exceeded(key string, now time.Time) bool {
	if strings.TrimSpace(key) == "" {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	item, exists := l.counter[key]
	if !exists || now.After(item.windowEnds) {
		delete(l.counter, key)
		return false
	}
	return item.count >= l.limit
}

func (l *localWindowLimiter) Reset(key string) {
	if strings.TrimSpace(key) == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.counter, key)
}

func (l *localWindowLimiter) Cleanup(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for key, item := range l.counter {
		if now.After(item.windowEnds) {
			delete(l.counter, key)
		}
	}
}

// RateLimit 对 /api/v1 请求做全局 IP 频率限制，默认 120 req/min。
func RateLimit() gin.HandlerFunc {
	limiter := newLocalWindowLimiter(defaultGlobalLimitPerMinute, defaultGlobalWindow)
	return func(c *gin.Context) {
		now := time.Now()
		clientIP := getRequestIP(c.Request)
		if !limiter.Allow(clientIP, now) {
			common.AbortFailMessage(c, http.StatusTooManyRequests, "too many requests, please try again later")
			return
		}

		// 惰性清理，减小长期运行的内存增长风险。
		if now.Unix()%30 == 0 {
			limiter.Cleanup(now)
		}
		c.Next()
	}
}

// LoginAttemptLimit 限制登录失败次数（IP + account），默认 8 次/15 分钟。
func LoginAttemptLimit() gin.HandlerFunc {
	limiter := newLocalWindowLimiter(defaultLoginFailLimit, defaultLoginFailWindow)
	return func(c *gin.Context) {
		now := time.Now()
		clientIP := getRequestIP(c.Request)
		account := extractAccount(c)
		failKey := buildLoginFailKey(clientIP, account)

		if strings.TrimSpace(account) != "" {
			if limiter.Exceeded(failKey, now) {
				common.AbortFailMessage(c, http.StatusTooManyRequests, "too many login attempts, please try again later")
				return
			}
		}

		c.Next()

		if strings.TrimSpace(account) == "" {
			return
		}
		switch c.Writer.Status() {
		case http.StatusUnauthorized:
			failCount := limiter.Increment(failKey, time.Now())
			if failCount >= defaultLoginFailLimit {
				klog.InfoS("Login attempts limited", "account", account, "clientIP", clientIP)
			}
		case http.StatusOK:
			limiter.Reset(failKey)
		}
	}
}

func buildLoginFailKey(clientIP, account string) string {
	normalizedIP := strings.TrimSpace(clientIP)
	if normalizedIP == "" {
		normalizedIP = "unknown"
	}
	normalizedAccount := strings.ToLower(strings.TrimSpace(account))
	return normalizedIP + "|" + normalizedAccount
}

func extractAccount(c *gin.Context) string {
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return ""
	}
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	var payload struct {
		Account string `json:"account"`
	}
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Account)
}

func getRequestIP(request *http.Request) string {
	if request == nil {
		return "unknown"
	}

	if forwardedFor := strings.TrimSpace(request.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		ip := strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
		if ip != "" {
			return ip
		}
	}
	if realIP := strings.TrimSpace(request.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host := strings.TrimSpace(request.RemoteAddr)
	if host == "" {
		return "unknown"
	}
	parsedHost, _, err := net.SplitHostPort(host)
	if err == nil && parsedHost != "" {
		return parsedHost
	}
	return host
}
