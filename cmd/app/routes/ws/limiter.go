package ws

import (
	"strings"
	"sync"
	"time"
)

// slidingWindowLimiter 与 middlewares.RateLimit 同构的滑动窗口计数，用于 WS 消息级限流。
type slidingWindowLimiter struct {
	mu      sync.Mutex
	counter map[string]windowCounter
	limit   int
	window  time.Duration
}

type windowCounter struct {
	count      int
	windowEnds time.Time
}

func newSlidingWindowLimiter(limit int, window time.Duration) *slidingWindowLimiter {
	if limit <= 0 {
		limit = 1
	}
	if window <= 0 {
		window = time.Second
	}
	return &slidingWindowLimiter{
		counter: make(map[string]windowCounter, 256),
		limit:   limit,
		window:  window,
	}
}

func (l *slidingWindowLimiter) allow(key string, now time.Time) bool {
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

func (l *slidingWindowLimiter) cleanup(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, item := range l.counter {
		if now.After(item.windowEnds) {
			delete(l.counter, k)
		}
	}
}
