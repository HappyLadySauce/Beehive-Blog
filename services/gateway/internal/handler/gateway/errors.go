package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	gwlogic "github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/logic/gateway"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func writeError(ctx context.Context, w http.ResponseWriter, err error) {
	requestID := requestIDFromContext(ctx)

	st, ok := status.FromError(err)
	if !ok {
		msg := sanitizeMessage(err)
		logx.WithContext(ctx).Errorf("request_failed request_id=%s code=INVALID_REQUEST http_status=%d error=%s",
			requestID, http.StatusBadRequest, msg)
		httpx.WriteJsonCtx(ctx, w, http.StatusBadRequest, map[string]any{
			"code":      "INVALID_REQUEST",
			"message":   msg,
			"requestId": requestID,
		})
		return
	}

	httpStatus, codeText := grpcCodeToHTTP(st.Code())
	msg := sanitizeMessage(st.Message())
	logx.WithContext(ctx).Errorf("request_failed request_id=%s code=%s grpc_code=%s http_status=%d error=%s",
		requestID, codeText, st.Code().String(), httpStatus, msg)
	httpx.WriteJsonCtx(ctx, w, httpStatus, map[string]any{
		"code":      codeText,
		"message":   msg,
		"requestId": requestID,
	})
}

func grpcCodeToHTTP(code codes.Code) (int, string) {
	switch code {
	case codes.Canceled:
		return http.StatusRequestTimeout, "REQUEST_CANCELED"
	case codes.InvalidArgument:
		return http.StatusBadRequest, "INVALID_ARGUMENT"
	case codes.OutOfRange:
		return http.StatusBadRequest, "OUT_OF_RANGE"
	case codes.FailedPrecondition:
		return http.StatusBadRequest, "FAILED_PRECONDITION"
	case codes.Aborted:
		return http.StatusConflict, "ABORTED"
	case codes.Unauthenticated:
		return http.StatusUnauthorized, "UNAUTHENTICATED"
	case codes.PermissionDenied:
		return http.StatusForbidden, "PERMISSION_DENIED"
	case codes.NotFound:
		return http.StatusNotFound, "NOT_FOUND"
	case codes.AlreadyExists:
		return http.StatusConflict, "ALREADY_EXISTS"
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests, "RATE_LIMITED"
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, "UPSTREAM_TIMEOUT"
	case codes.Unavailable:
		return http.StatusServiceUnavailable, "UPSTREAM_UNAVAILABLE"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}

func sanitizeMessage(v any) string {
	msg := strings.TrimSpace(fmt.Sprint(v))
	if msg == "" || msg == "<nil>" {
		return "unknown error"
	}
	return msg
}

func requestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v := ctx.Value(gwlogic.RequestIDContextKey)
	requestID, _ := v.(string)
	return requestID
}
