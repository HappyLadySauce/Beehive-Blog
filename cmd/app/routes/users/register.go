package users

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/passwd"
	authsession "github.com/HappyLadySauce/Beehive-Blog/pkg/auth/session"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// Register creates a new user and issues JWT credentials (auto-login).
// Avatar binding is intentionally omitted here (attachments lack ownership columns); bind avatars after authenticated flows.
// Register 创建新用户并签发 JWT 凭证（自动登录）。
// 头像绑定有意不在此完成（附件表无归属列）；请在登录态流程后再绑定头像。
func (u *UsersController) register(ctx *gin.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
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

	meta := authsession.ClientMeta{
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	}

	var pair jwt.TokenPair

	err = u.svc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return mapRegisterUniqueViolation(pgErr)
			}
			return fmt.Errorf("create user: %w", err)
		}

		cred := model.UserCredential{
			UserID:       user.ID,
			PasswordHash: hash,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := tx.Create(&cred).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return mapRegisterUniqueViolation(pgErr)
			}
			return fmt.Errorf("create credential: %w", err)
		}

		p, _, err := authsession.IssuePairInTx(tx, u.svc.Token, &user, meta)
		if err != nil {
			return fmt.Errorf("issue token pair: %w", err)
		}
		pair = p
		return nil
	})
	if err != nil {
		var appErr *common.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, common.NewInternal("failed to register user", err)
	}

	klog.InfoS("User registered", "uid", user.ID, "username", user.Username)

	return &v1.RegisterResponse{
		Token: v1.AuthToken{
			AccessToken:  pair.Access.Token,
			TokenType:    pair.TokenType,
			ExpiresIn:    pair.Access.ExpiresIn,
			RefreshToken: pair.Refresh.Token,
		},
	}, nil
}

func mapRegisterUniqueViolation(pgErr *pgconn.PgError) *common.AppError {
	// Map stable constraint names to safe API-facing conflict messages (Err left nil to skip noisy 4xx logs).
	// 将稳定约束名映射为安全的冲突文案（Err 置 nil，避免 4xx 路径触发冗余 Error 日志）。
	switch pgErr.ConstraintName {
	case "ux_identity_users_username":
		return common.NewConflict("username is already taken", nil)
	case "ux_identity_users_email":
		return common.NewConflict("email is already registered", nil)
	default:
		return common.NewConflict("registration conflict", nil)
	}
}

// Register is the Gin entrypoint for POST /api/v1/users/register (JSON bind + Swagger).
// Register 为 POST /api/v1/users/register 的 Gin 入口（JSON 绑定与 Swagger 元数据）。
//
// @Summary      Register a new user
// @Description  Creates identity.users plus credentials and returns tokens (auto-login). Avatar binding is deferred until authenticated flows. 中文：创建用户与凭证并返回令牌（自动登录）；头像请在登录态流程中再绑定。
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      v1.RegisterRequest  true  "Registration payload"
// @Success      200   {object}  common.BaseResponse{data=v1.RegisterResponse}  "Issued access and refresh tokens"
// @Failure      400   {object}  common.BaseResponse                            "Validation error"
// @Failure      409   {object}  common.BaseResponse                            "Username or email conflict"
// @Failure      500   {object}  common.BaseResponse                            "Internal error"
// @Router       /api/v1/users/register [post]
func (u *UsersController) Register(ctx *gin.Context) {
	var req v1.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := u.register(ctx, &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
