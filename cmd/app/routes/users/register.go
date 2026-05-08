package users

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/passwd"
	authsession "github.com/HappyLadySauce/Beehive-Blog/pkg/auth/session"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// Register creates a new user and issues JWT credentials (auto-login).
// Register 创建新用户并签发 JWT 凭证（自动登录）。
func (u *UsersController) Register(ctx *gin.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	// Check username uniqueness among live rows.
	// 检查用户名在活跃行中的唯一性。
	var existing model.User
	if err := u.svc.DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, common.NewConflict("username is already taken", nil)
	} else if err != gorm.ErrRecordNotFound {
		return nil, common.NewInternal("failed to register user", fmt.Errorf("check username: %w", err))
	}

	// Check email uniqueness if provided.
	// 如提供邮箱则检查其唯一性。
	if req.Email != "" {
		if err := u.svc.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
			return nil, common.NewConflict("email is already registered", nil)
		} else if err != gorm.ErrRecordNotFound {
			return nil, common.NewInternal("failed to register user", fmt.Errorf("check email: %w", err))
		}
	}

	// Hash the password with bcrypt.
	// 用 bcrypt 哈希密码。
	hash, err := passwd.Hash(req.Password)
	if err != nil {
		return nil, common.NewInternal("failed to register user", fmt.Errorf("hash password: %w", err))
	}

	// Build the user row; pointer fields default to nil when empty.
	// 构造用户行；指针字段为空时即为 nil。
	now := time.Now()
	user := model.User{
		Username:  req.Username,
		Role:      "member",
		Status:    "active",
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
			PasswordHash: hash,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := tx.Create(&cred).Error; err != nil {
			return fmt.Errorf("create credential: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, common.NewInternal("failed to register user", err)
	}

	klog.InfoS("User registered", "uid", user.ID, "username", user.Username)

	// Issue a session-bound token pair for auto-login after registration.
	// 签发绑定会话的令牌对，实现注册后自动登录。
	pair, _, err := authsession.IssuePair(u.svc.DB, u.svc.Token, &user, authsession.ClientMeta{
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	})
	if err != nil {
		return nil, common.NewInternal("failed to issue token", err)
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
