package users

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// list queries users with pagination and optional filters.
// list 查询用户列表（分页+可选筛选）。
func (u *UsersController) list(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	query := u.svc.DB.WithContext(ctx).Model(&model.User{})
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}
	if req.Search != "" {
		pattern := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(CAST(email AS text)) LIKE ?", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, common.NewInternal("failed to list users", fmt.Errorf("count: %w", err))
	}

	var users []model.User
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, common.NewInternal("failed to list users", fmt.Errorf("find: %w", err))
	}

	items := make([]v1.UserItem, len(users))
	for i, usr := range users {
		items[i] = toUserItem(usr)
	}

	return &v1.ListUsersResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// List handles GET /api/v1/users (admin).
// List 处理 GET /api/v1/users（管理员）。
//
//	@Summary		List users
//	@Description	Paginated list with optional status, role and search filters. 中文：分页用户列表，支持状态、角色和搜索筛选。
//	@Tags			users
//	@Security		BearerAuth
//	@Produce		json
//	@Param			page		query		int		false	"Page number (default 1)"	default(1)
//	@Param			page_size	query		int		false	"Items per page (default 20, max 100)"	default(20)
//	@Param			status		query		string	false	"Filter by status"	Enums(active, disabled, locked, pending)
//	@Param			role		query		string	false	"Filter by role"	Enums(member, admin)
//	@Param			search		query		string	false	"Search username or email"
//	@Success		200			{object}	common.BaseResponse{data=v1.ListUsersResponse}
//	@Failure		401			{object}	common.BaseResponse
//	@Failure		403			{object}	common.BaseResponse
//	@Router			/api/v1/users [get]
func (u *UsersController) List(ctx *gin.Context) {
	var req v1.ListUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid query parameters", err))
		return
	}
	resp, err := u.list(ctx.Request.Context(), &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
