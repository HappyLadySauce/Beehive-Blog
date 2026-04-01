package public

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	authutil "github.com/HappyLadySauce/Beehive-Blog/pkg/utils/auth"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/passwd"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// Register 用户注册
func (s *PublicService) Register(ctx context.Context, spec *v1.RegisterRequest, request *http.Request) (*v1.RegisterResponse, int, error) {
	if spec == nil || request == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	// 使用事务处理
	tx := s.svc.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		klog.ErrorS(tx.Error, "Failed to begin transaction")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	// 确保事务最终提交或回滚
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 1. 检查用户名是否已存在
	var existingUser models.User
	if err := tx.Where("username = ?", spec.Username).First(&existingUser).Error; err == nil {
		tx.Rollback()
		klog.InfoS("Username already exists", "username", spec.Username)
		return nil, http.StatusConflict, errors.New("registration information is invalid or already in use")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		klog.ErrorS(err, "Failed to check username", "username", spec.Username)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	// 2. 检查邮箱是否已存在
	if err := tx.Where("email = ?", spec.Email).First(&existingUser).Error; err == nil {
		tx.Rollback()
		klog.InfoS("Email already exists", "email", spec.Email)
		return nil, http.StatusConflict, errors.New("registration information is invalid or already in use")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		klog.ErrorS(err, "Failed to check email", "email", spec.Email)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	// 3. 加密密码
	hashedPassword, err := passwd.HashPassword(spec.Password)
	if err != nil {
		tx.Rollback()
		klog.ErrorS(err, "Failed to hash password")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	// 4. 创建用户
	user := &models.User{
		Username: spec.Username,
		Email:    spec.Email,
		Password: hashedPassword,
		Role:     models.UserRoleUser,
		Status:   models.UserStatusActive,
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		klog.ErrorS(err, "Failed to create user", "username", spec.Username)
		return nil, http.StatusInternalServerError, errors.New("failed to create user")
	}

	// 5. 提交事务
	if err := tx.Commit().Error; err != nil {
		klog.ErrorS(err, "Failed to commit transaction")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	// 6. 生成 JWT Token
	jwtSecret := s.svc.Config.JWTOptions.JWTSecret
	if strings.TrimSpace(jwtSecret) == "" {
		klog.Error("JWTSecret is not configured")
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}

	// 纯 Redis 鉴权模式下，注册后必须回写鉴权快照，否则新签发 token 无法通过中间件校验。
	if s.svc.Redis == nil {
		klog.Error("Redis client is not configured")
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
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

	tokenPair, err := jwt.GenerateToken(
		jwtSecret,
		user.ID,
		user.Username,
		string(user.Role),
		s.svc.Config.JWTOptions.ExpireDuration,
		s.svc.Config.JWTOptions.RefreshTokenExpireDuration,
	)
	if err != nil {
		klog.ErrorS(err, "Failed to generate token", "userID", user.ID)
		return nil, http.StatusInternalServerError, errors.New("failed to generate token")
	}

	klog.InfoS("User registered successfully",
		"userID", user.ID,
		"username", user.Username,
		"email", user.Email,
		"clientIP", request.RemoteAddr,
	)

	return &v1.RegisterResponse{
		Token:        tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, http.StatusOK, nil
}
