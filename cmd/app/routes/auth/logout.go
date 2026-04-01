package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/jwt"
	"k8s.io/klog/v2"
)

const logoutBearerPrefix = "Bearer "

func (s *AuthService) Logout(ctx context.Context, spec *v1.LogoutRequest, request *http.Request) (*v1.LogoutResponse, int, error) {
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

	rawToken, err := extractLogoutBearerToken(request.Header.Get("Authorization"))
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("invalid authorization header")
	}
	claims, err := jwt.ParseToken(jwtSecret, rawToken)
	if err != nil || claims == nil || claims.UserID <= 0 {
		return nil, http.StatusUnauthorized, errors.New("invalid or expired token")
	}

	authCacheKey := fmt.Sprintf("auth:user:%d", claims.UserID)
	if err := s.svc.Redis.Del(ctx, authCacheKey).Err(); err != nil {
		klog.ErrorS(err, "Failed to remove auth snapshot from redis", "userID", claims.UserID)
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}

	klog.InfoS("User logout successfully", "userID", claims.UserID, "clientIP", request.RemoteAddr)
	return &v1.LogoutResponse{
		Message: "logout success",
	}, http.StatusOK, nil
}

func extractLogoutBearerToken(authHeader string) (string, error) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" || !strings.HasPrefix(authHeader, logoutBearerPrefix) {
		return "", errors.New("missing bearer token")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, logoutBearerPrefix))
	if token == "" {
		return "", errors.New("empty token")
	}
	return token, nil
}
