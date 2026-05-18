package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/passwd"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// update applies partial changes to an existing user.
// update 对已有用户应用部分修改。
func (u *UsersController) update(ctx context.Context, id int64, req *v1.UpdateUserRequest) (*v1.UserDetailResponse, error) {
	var user model.User
	if err := u.svc.DB.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewNotFound("user not found", nil)
		}
		return nil, common.NewInternal("failed to update user", fmt.Errorf("first: %w", err))
	}

	updates := map[string]interface{}{"updated_at": time.Now()}

	if req.Username != nil && *req.Username != user.Username {
		reject := u.checkUsernameConflict(ctx, *req.Username, id)
		if reject != nil {
			return nil, reject
		}
		updates["username"] = *req.Username
	}

	if req.Email != nil {
		if *req.Email == "" {
			updates["email"] = nil
		} else if user.Email == nil || *req.Email != *user.Email {
			reject := u.checkEmailConflict(ctx, *req.Email, id)
			if reject != nil {
				return nil, reject
			}
			updates["email"] = *req.Email
		}
	}

	if req.Nickname != nil {
		if *req.Nickname == "" {
			updates["nickname"] = nil
		} else {
			updates["nickname"] = *req.Nickname
		}
	}
	if req.Phone != nil {
		if *req.Phone == "" {
			updates["phone"] = nil
		} else {
			updates["phone"] = *req.Phone
		}
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.AvatarAttachmentID != nil {
		if *req.AvatarAttachmentID == 0 {
			updates["avatar_attachment_id"] = nil
		} else {
			var attachment model.Attachment
			err := u.svc.DB.WithContext(ctx).
				Where("id = ? AND purpose = ? AND upload_status = ? AND status = ?", *req.AvatarAttachmentID, "avatar", "ready", "active").
				First(&attachment).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, common.NewBadRequest("avatar attachment is not available", nil)
			}
			if err != nil {
				return nil, common.NewInternal("failed to validate avatar attachment", fmt.Errorf("avatar lookup: %w", err))
			}
			updates["avatar_attachment_id"] = *req.AvatarAttachmentID
		}
	}

	err := u.svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			return mapCrudUniqueViolation(err)
		}
		if req.Password != nil && *req.Password != "" {
			hash, err := passwd.Hash(*req.Password)
			if err != nil {
				return common.NewInternal("failed to update user", fmt.Errorf("hash password: %w", err))
			}
			now := time.Now()
			var cred model.UserCredential
			err = tx.Where("user_id = ? AND deleted_at IS NULL", id).First(&cred).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				cred = model.UserCredential{
					UserID:       id,
					PasswordHash: hash,
					CreatedAt:    now,
					UpdatedAt:    now,
				}
				if err := tx.Create(&cred).Error; err != nil {
					return fmt.Errorf("create credential: %w", err)
				}
			} else if err == nil {
				if err := tx.Model(&cred).Where("user_id = ? AND deleted_at IS NULL", id).
					Updates(map[string]interface{}{"password_hash": hash, "updated_at": now}).Error; err != nil {
					return fmt.Errorf("update credential: %w", err)
				}
			} else {
				return fmt.Errorf("find credential: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, mapCrudError(err)
	}

	var updated model.User
	if err := u.svc.DB.WithContext(ctx).First(&updated, id).Error; err != nil {
		return nil, common.NewInternal("failed to read updated user", fmt.Errorf("re-fetch: %w", err))
	}

	return toUserDetail(updated), nil
}

// Update handles PATCH /api/v1/users/:id (admin).
// Update 处理 PATCH /api/v1/users/:id（管理员）。
//
//	@Summary		Update user
//	@Description	Partially update user fields. Pointer fields: nil=unchanged, ""=clear, "v"=set. 中文：部分更新用户字段。指针字段：nil=不修改，空串=清空，有值=设置。
//	@Tags			users
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"User ID"
//	@Param			body	body		v1.UpdateUserRequest	true	"Update user payload"
//	@Success		200		{object}	common.BaseResponse{data=v1.UpdateUserResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Router			/api/v1/users/{id} [patch]
func (u *UsersController) Update(ctx *gin.Context) {
	id, ok := parseUserID(ctx)
	if !ok {
		return
	}
	var req v1.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := u.update(ctx.Request.Context(), id, &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
