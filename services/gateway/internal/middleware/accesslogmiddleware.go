package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/config"
	gwlogic "github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/logic/gateway"
	"github.com/zeromicro/go-zero/core/logx"
)

type AccessLogMiddleware struct {
	slowRequestWarnThresholdMs int64
}

const defaultSlowRequestWarnThresholdMs int64 = 500

func NewAccessLogMiddleware(conf config.AccessLogConf) *AccessLogMiddleware {
	threshold := conf.SlowRequestWarnThresholdMs
	if threshold <= 0 {
		threshold = defaultSlowRequestWarnThresholdMs
	}
	return &AccessLogMiddleware{
		slowRequestWarnThresholdMs: threshold,
	}
}

func (m *AccessLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next(recorder, r)

		ctx := r.Context()
		latencyMs := time.Since(startedAt).Milliseconds()
		logger := logx.WithContext(ctx)
		if latencyMs > m.slowRequestWarnThresholdMs {
			logger.Slowf(
				"access_slow method=%s path=%s status=%d latency_ms=%d threshold_ms=%d request_id=%s user_id=%d",
				r.Method,
				r.URL.Path,
				recorder.statusCode,
				latencyMs,
				m.slowRequestWarnThresholdMs,
				accessRequestIDFromContext(ctx),
				accessUserIDFromContext(ctx),
			)
			return
		}

		logger.Infof(
			"access method=%s path=%s status=%d latency_ms=%d request_id=%s user_id=%d",
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			latencyMs,
			accessRequestIDFromContext(ctx),
			accessUserIDFromContext(ctx),
		)
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func accessRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v := ctx.Value(gwlogic.RequestIDContextKey)
	requestID, _ := v.(string)
	return requestID
}

func accessUserIDFromContext(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	v := ctx.Value(gwlogic.AuthUserIDContextKey)
	userID, _ := v.(int64)
	return userID
}
