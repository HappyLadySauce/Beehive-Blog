package public

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	authutil "github.com/HappyLadySauce/Beehive-Blog/pkg/utils/auth"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/passwd"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// PublicService handles unauthenticated /public endpoints business logic (e.g. login).
type PublicService struct {
	svc *svc.ServiceContext
}

// NewPublicService constructs a PublicService.
func NewPublicService(svc *svc.ServiceContext) *PublicService {
	return &PublicService{svc: svc}
}

// Login authenticates by username or email and returns tokens.
func (s *PublicService) Login(ctx context.Context, spec *v1.LoginRequest, request *http.Request) (*v1.LoginResponse, int, error) {
	if spec == nil {
		return nil, http.StatusBadRequest, errors.New("invalid login request")
	}

	account := strings.TrimSpace(spec.Account)
	if account == "" {
		return nil, http.StatusBadRequest, errors.New("account is required")
	}

	if strings.TrimSpace(s.svc.Config.JWTOptions.JWTSecret) == "" {
		klog.Error("JWTSecret is not configured")
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}
	if s.svc.Redis == nil {
		klog.Error("Redis client is not configured")
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}

	var user models.User
	query := s.svc.DB.WithContext(ctx)
	if strings.Contains(account, "@") {
		query = query.Where("email = ?", account)
	} else {
		query = query.Where("username = ?", account)
	}

	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusUnauthorized, errors.New("invalid account or password")
		}
		klog.ErrorS(err, "Failed to query user for login", "account", account)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	if !passwd.VerifyPassword(spec.Password, user.Password) {
		return nil, http.StatusUnauthorized, errors.New("invalid account or password")
	}

	if user.Status != models.UserStatusActive {
		return nil, http.StatusForbidden, errors.New("user is not active")
	}

	tokenPair, err := jwt.GenerateToken(
		s.svc.Config.JWTOptions.JWTSecret,
		user.ID,
		user.Username,
		string(user.Role),
		s.svc.Config.JWTOptions.ExpireDuration,
		s.svc.Config.JWTOptions.RefreshTokenExpireDuration,
	)
	if err != nil {
		klog.ErrorS(err, "Failed to generate login token", "userID", user.ID)
		return nil, http.StatusInternalServerError, errors.New("failed to generate token")
	}

	loginAt := time.Now()
	if err := s.svc.DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", user.ID).Update("last_login_at", loginAt).Error; err != nil {
		klog.ErrorS(err, "Failed to update last login time", "userID", user.ID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	authCacheKey := authutil.UserAuthCacheKey(user.ID)
	if err := s.svc.Redis.HSet(ctx, authCacheKey, map[string]interface{}{
		"role":   string(user.Role),
		"status": string(user.Status),
	}).Err(); err != nil {
		klog.ErrorS(err, "Failed to write auth snapshot to redis", "userID", user.ID)
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}
	if err := s.svc.Redis.Expire(ctx, authCacheKey, s.svc.Config.JWTOptions.ExpireDuration).Err(); err != nil {
		klog.ErrorS(err, "Failed to set auth snapshot ttl", "userID", user.ID)
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}

	klog.InfoS("User login successfully",
		"userID", user.ID,
		"username", user.Username,
		"clientIP", request.RemoteAddr,
	)

	return &v1.LoginResponse{
		Token:        tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, http.StatusOK, nil
}
