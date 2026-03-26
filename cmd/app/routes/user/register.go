package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/passwd"
	"k8s.io/klog/v2"
)

// Register 用户注册
func (s *UserService) Register(ctx context.Context, spec *v1.RegisterRequest, request *http.Request) (*v1.RegisterResponse, int, error) {
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
		return nil, http.StatusConflict, errors.New("username already exists")
	} else if !errors.Is(err, errors.New("record not found")) {
		tx.Rollback()
		klog.ErrorS(err, "Failed to check username", "username", spec.Username)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	// 2. 检查邮箱是否已存在
	if err := tx.Where("email = ?", spec.Email).First(&existingUser).Error; err == nil {
		tx.Rollback()
		klog.InfoS("Email already exists", "email", spec.Email)
		return nil, http.StatusConflict, errors.New("email already exists")
	} else if !errors.Is(err, errors.New("record not found")) {
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
	// 从配置中获取 JWT Secret
	jwtSecret := s.svc.Config.JWTOptions.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "beehive-blog-default-secret-key-change-in-production"
		klog.Warning("JWTSecret is not set, using default secret. Please set a secure secret in production!")
	}

	tokenPair, err := jwt.GenerateToken(jwtSecret, user.ID, user.Username, string(user.Role), s.svc.Config.JWTOptions.ExpireDuration, s.svc.Config.JWTOptions.RefreshTokenExpireDuration)
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
