package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// get fetches a single user by ID.
// get 根据 ID 获取单个用户。
func (u *UsersController) get(ctx context.Context, id int64) (*v1.UserDetailResponse, error) {
	var user model.User
	if err := u.svc.DB.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewNotFound("user not found", nil)
		}
		return nil, common.NewInternal("failed to get user", fmt.Errorf("first: %w", err))
	}
	return toUserDetail(user), nil
}

// Get handles GET /api/v1/users/:id (admin).
// Get 处理 GET /api/v1/users/:id（管理员）。
//
//	@Summary		Get user detail
//	@Description	Returns full user detail by ID. 中文：根据 ID 返回用户完整详情。
//	@Tags			users
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.UserDetailResponse}
//	@Failure		404	{object}	common.BaseResponse
//	@Router			/api/v1/users/{id} [get]
func (u *UsersController) Get(ctx *gin.Context) {
	id, ok := parseUserID(ctx)
	if !ok {
		return
	}
	resp, err := u.get(ctx.Request.Context(), id)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
