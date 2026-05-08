package users

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// bcryptCost balances security and server load; 12 is current industry default.
// bcryptCost 平衡安全性与服务器负载；12 为当前业界默认值。
const bcryptCost = 12

// Register creates a new user and issues JWT credentials (auto-login).
// Register 创建新用户并签发 JWT 凭证（自动登录）。
func (u *UsersController) Register(ctx *gin.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	// Check username uniqueness among live rows.
	// 检查用户名在活跃行中的唯一性。
	var existing model.User
	if err := u.svc.DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("username %q is already taken", req.Username)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("check username: %w", err)
	}

	// Check email uniqueness if provided.
	// 如提供邮箱则检查其唯一性。
	if req.Email != "" {
		if err := u.svc.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
			return nil, fmt.Errorf("email %q is already registered", req.Email)
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("check email: %w", err)
		}
	}

	// Hash the password with bcrypt.
	// 用 bcrypt 哈希密码。
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Build the user row; pointer fields default to nil when empty.
	// 构造用户行；指针字段为空时即为 nil。
	now := time.Now()
	user := model.User{
		Username: req.Username,
		Role:     "member",
		Status:   "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if req.Email != "" {
		user.Email = &req.Email
	}
	if req.Nickname != "" {
		user.Nickname = &req.Nickname
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if req.AvatarAttachmentID != nil {
		user.AvatarAttachmentID = req.AvatarAttachmentID
	}

	// Create user and credential in a transaction.
	// 在事务中创建用户和凭证。
	err = u.svc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		cred := model.UserCredential{
			UserID:       user.ID,
			PasswordHash: string(hash),
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := tx.Create(&cred).Error; err != nil {
			return fmt.Errorf("create credential: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	klog.InfoS("User registered", "uid", user.ID, "username", user.Username)

	// Issue JWT token pair for auto-login after registration.
	// 签发 JWT 令牌对，实现注册后自动登录。
	pair, err := u.svc.Token.IssuePair(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("issue JWT: %w", err)
	}

	return &v1.RegisterResponse{
		Token: v1.AuthToken{
			AccessToken:  pair.Access.Token,
			TokenType:    pair.TokenType,
			ExpiresIn:    pair.Access.ExpiresIn,
			RefreshToken: pair.Refresh.Token,
		},
	}, nil
}
