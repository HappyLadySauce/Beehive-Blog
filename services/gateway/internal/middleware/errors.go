package middleware

import (
	"context"
	"net/http"

	gwlogic "github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/logic/gateway"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func writeJSONError(ctx context.Context, w http.ResponseWriter, httpStatus int, code, message string) {
	requestID := requestIDFromContext(ctx)
	logx.WithContext(ctx).Errorf("request_failed request_id=%s code=%s http_status=%d error=%s",
		requestID, code, httpStatus, message)
	httpx.WriteJsonCtx(ctx, w, httpStatus, map[string]any{
		"code":      code,
		"message":   message,
		"requestId": requestID,
	})
}

func requestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v := ctx.Value(gwlogic.RequestIDContextKey)
	requestID, _ := v.(string)
	return requestID
}
