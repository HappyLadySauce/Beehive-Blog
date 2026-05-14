package users

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/passwd"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// create inserts a new user and optionally a password credential.
// create 插入新用户，可附带密码凭证。
func (u *UsersController) create(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	reject := u.checkUsernameConflict(ctx, req.Username, 0)
	if reject != nil {
		return nil, reject
	}

	if req.Email != nil && *req.Email != "" {
		reject = u.checkEmailConflict(ctx, *req.Email, 0)
		if reject != nil {
			return nil, reject
		}
	}

	var hash string
	hasPassword := req.Password != nil && *req.Password != ""
	if hasPassword {
		var err error
		hash, err = passwd.Hash(*req.Password)
		if err != nil {
			return nil, common.NewInternal("failed to create user", fmt.Errorf("hash password: %w", err))
		}
	}

	now := time.Now()
	user := model.User{
		Username:  req.Username,
		Role:      "member",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if req.Email != nil {
		user.Email = req.Email
	}
	if req.Nickname != nil {
		user.Nickname = req.Nickname
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Role != nil && *req.Role != "" {
		user.Role = *req.Role
	}
	if req.Status != nil && *req.Status != "" {
		user.Status = *req.Status
	}

	err := u.svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return mapCrudUniqueViolation(err)
		}
		if hasPassword {
			cred := model.UserCredential{
				UserID:       user.ID,
				PasswordHash: hash,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := tx.Create(&cred).Error; err != nil {
				return fmt.Errorf("create credential: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, mapCrudError(err)
	}

	klog.InfoS("Admin created user", "uid", user.ID, "username", user.Username)
	return &v1.CreateUserResponse{ID: user.ID}, nil
}

// Create handles POST /api/v1/users (admin).
// Create 处理 POST /api/v1/users（管理员）。
//
//	@Summary		Create a user
//	@Description	Admin creates a user with optional password. 中文：管理员创建用户，可附带密码。
//	@Tags			users
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		v1.CreateUserRequest	true	"Create user payload"
//	@Success		200		{object}	common.BaseResponse{data=v1.CreateUserResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Router			/api/v1/users [post]
func (u *UsersController) Create(ctx *gin.Context) {
	var req v1.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := u.create(ctx.Request.Context(), &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
