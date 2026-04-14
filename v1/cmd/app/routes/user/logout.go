package user

import (
	"context"
	"errors"
	"net/http"
	"strings"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	authutil "github.com/HappyLadySauce/Beehive-Blog/pkg/utils/auth"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/jwt"
	"k8s.io/klog/v2"
)

// Logout invalidates the current session snapshot in Redis for the bearer token subject.
func (s *UserService) Logout(ctx context.Context, spec *v1.LogoutRequest, request *http.Request) (*v1.LogoutResponse, int, error) {
	if spec == nil {
		return nil, http.StatusBadRequest, errors.New("invalid logout request")
	}
	if request == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	if s.svc.Redis == nil {
		klog.Error("Redis client is not configured")
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}

	jwtSecret := strings.TrimSpace(s.svc.Config.JWTOptions.JWTSecret)
	if jwtSecret == "" {
		klog.Error("JWTSecret is not configured")
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}

	rawToken, err := authutil.ExtractBearerToken(request.Header.Get("Authorization"))
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("invalid authorization header")
	}
	claims, err := jwt.ParseToken(jwtSecret, rawToken)
	if err != nil || claims == nil || claims.UserID <= 0 {
		return nil, http.StatusUnauthorized, errors.New("invalid or expired token")
	}

	authCacheKey := authutil.UserAuthCacheKey(claims.UserID)
	if err := s.svc.Redis.Del(ctx, authCacheKey).Err(); err != nil {
		klog.ErrorS(err, "Failed to remove auth snapshot from redis", "userID", claims.UserID)
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}

	klog.InfoS("User logout successfully", "userID", claims.UserID, "clientIP", request.RemoteAddr)
	return &v1.LogoutResponse{
		Message: "logout success",
	}, http.StatusOK, nil
}
