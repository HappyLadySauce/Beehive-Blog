package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

const adminDeletedUserRevokedReason = "admin_deleted_user"

// delete soft-deletes a user by ID.
// delete 根据 ID 软删除用户。
func (u *UsersController) del(ctx context.Context, id int64) error {
	var user model.User
	if err := u.svc.DB.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewNotFound("user not found", nil)
		}
		return common.NewInternal("failed to delete user", fmt.Errorf("first: %w", err))
	}

	now := time.Now()
	err := u.svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&model.User{}).Error; err != nil {
			return fmt.Errorf("delete user: %w", err)
		}
		if err := tx.Where("user_id = ?", id).Delete(&model.UserCredential{}).Error; err != nil {
			return fmt.Errorf("delete credentials: %w", err)
		}
		if err := tx.Where("user_id = ?", id).Delete(&model.UserIdentity{}).Error; err != nil {
			return fmt.Errorf("delete identities: %w", err)
		}
		if err := tx.Model(&model.UserSession{}).
			Where("user_id = ? AND revoked_at IS NULL AND rotated_at IS NULL AND expires_at > ?", id, now).
			Updates(map[string]interface{}{
				"revoked_at":     now,
				"revoked_reason": adminDeletedUserRevokedReason,
				"updated_at":     now,
			}).Error; err != nil {
			return fmt.Errorf("revoke sessions: %w", err)
		}
		return nil
	})
	if err != nil {
		return common.NewInternal("failed to delete user", err)
	}

	klog.InfoS("Admin deleted user", "uid", user.ID, "username", user.Username)
	return nil
}

// Delete handles DELETE /api/v1/users/:id (admin).
// Delete 处理 DELETE /api/v1/users/:id（管理员）。
//
//	@Summary		Delete user
//	@Description	Soft-deletes a user by ID. 中文：根据 ID 软删除用户。
//	@Tags			users
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Router			/api/v1/users/{id} [delete]
func (u *UsersController) Delete(ctx *gin.Context) {
	id, ok := parseUserID(ctx)
	if !ok {
		return
	}
	if err := u.del(ctx.Request.Context(), id); err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, gin.H{})
}
