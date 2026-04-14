package gateway

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func writeError(ctx context.Context, w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		httpx.ErrorCtx(ctx, w, err)
		return
	}

	httpStatus, codeText := grpcCodeToHTTP(st.Code())
	httpx.WriteJsonCtx(ctx, w, httpStatus, map[string]any{
		"code":    codeText,
		"message": st.Message(),
	})
}

func grpcCodeToHTTP(code codes.Code) (int, string) {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest, "INVALID_ARGUMENT"
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
