package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	authutil "github.com/HappyLadySauce/Beehive-Blog/v1/pkg/utils/auth"
	"github.com/HappyLadySauce/Beehive-Blog/v1/pkg/utils/jwt"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// Refresh exchanges a valid refresh token for a new token pair.
//
// Flow:
//  1. Parse & verify the refresh token signature and expiry.
//  2. Extract userID from claims.Subject.
//  3. Load user from DB; verify status = active.
//  4. Issue a new TokenPair.
//  5. Overwrite Redis auth snapshot (HSet + Expire) to re-anchor the session TTL.
func (s *Service) Refresh(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.RefreshTokenResponse, int, error) {
	if req == nil {
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

	claims, err := jwt.ParseRefreshToken(jwtSecret, req.RefreshToken)
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("invalid or expired refresh token")
	}

	subject := strings.TrimSpace(claims.Subject)
	if subject == "" {
		return nil, http.StatusUnauthorized, errors.New("invalid refresh token claims")
	}
	userID, err := strconv.ParseInt(subject, 10, 64)
	if err != nil || userID <= 0 {
		return nil, http.StatusUnauthorized, errors.New("invalid refresh token claims")
	}

	var user models.User
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusUnauthorized, errors.New("user not found")
		}
		klog.ErrorS(err, "Refresh: load user", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if user.Status != models.UserStatusActive {
		return nil, http.StatusForbidden, errors.New("user account is not active")
	}

	tokenPair, err := jwt.GenerateToken(
		jwtSecret,
		user.ID,
		user.Username,
		string(user.Role),
		s.svc.Config.JWTOptions.ExpireDuration,
		s.svc.Config.JWTOptions.RefreshTokenExpireDuration,
	)
	if err != nil {
		klog.ErrorS(err, "Refresh: generate token", "userID", user.ID)
		return nil, http.StatusInternalServerError, errors.New("failed to generate token")
	}

	authCacheKey := authutil.UserAuthCacheKey(user.ID)
	if err := s.svc.Redis.HSet(ctx, authCacheKey, map[string]interface{}{
		"role":   string(user.Role),
		"status": string(user.Status),
	}).Err(); err != nil {
		klog.ErrorS(err, "Refresh: write auth snapshot", "userID", user.ID)
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}
	if err := s.svc.Redis.Expire(ctx, authCacheKey, s.svc.Config.JWTOptions.ExpireDuration).Err(); err != nil {
		klog.ErrorS(err, "Refresh: set auth snapshot TTL", "userID", user.ID)
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}

	klog.InfoS("Token refreshed", "userID", user.ID)
	return &v1.RefreshTokenResponse{
		Token:        tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, http.StatusOK, nil
}
