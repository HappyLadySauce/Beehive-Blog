package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// parseUserID extracts and validates an int64 user ID from the :id path parameter.
// parseUserID 从 :id 路径参数提取并校验 int64 用户 ID。
func parseUserID(ctx *gin.Context) (int64, bool) {
	var uri struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	if err := ctx.ShouldBindUri(&uri); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid user id", err))
		return 0, false
	}
	return uri.ID, true
}

func (u *UsersController) checkUsernameConflict(ctx context.Context, username string, excludeID int64) *common.AppError {
	var existing model.User
	err := u.svc.DB.WithContext(ctx).Where("username = ? AND id != ?", username, excludeID).First(&existing).Error
	if err == nil {
		return common.NewConflict("username is already taken", nil)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return common.NewInternal("failed to check username uniqueness", fmt.Errorf("username check: %w", err))
	}
	return nil
}

func (u *UsersController) checkEmailConflict(ctx context.Context, email string, excludeID int64) *common.AppError {
	var existing model.User
	err := u.svc.DB.WithContext(ctx).Where("email = ? AND id != ?", email, excludeID).First(&existing).Error
	if err == nil {
		return common.NewConflict("email is already registered", nil)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return common.NewInternal("failed to check email uniqueness", fmt.Errorf("email check: %w", err))
	}
	return nil
}

func toUserItem(u model.User) v1.UserItem {
	return v1.UserItem{
		ID:                 u.ID,
		Username:           u.Username,
		Email:              u.Email,
		Nickname:           u.Nickname,
		Phone:              u.Phone,
		AvatarAttachmentID: u.AvatarAttachmentID,
		Role:               u.Role,
		Status:             u.Status,
		LastLoginAt:        u.LastLoginAt,
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
	}
}

func toUserDetail(u model.User) *v1.UserDetailResponse {
	return &v1.UserDetailResponse{
		UserItem: toUserItem(u),
	}
}

func mapCrudUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "ux_identity_users_username":
			return common.NewConflict("username is already taken", nil)
		case "ux_identity_users_email":
			return common.NewConflict("email is already registered", nil)
		default:
			return common.NewConflict("conflict", nil)
		}
	}
	return err
}

func mapCrudError(err error) error {
	var appErr *common.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return common.NewInternal("operation failed", err)
}
