package tags

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// list queries tags with pagination and optional filters.
// list 查询标签列表（分页+可选筛选）。
func (t *TagsController) list(ctx context.Context, req *v1.ListTagsRequest, admin bool) (*v1.ListTagsResponse, error) {
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	query := t.svc.DB.WithContext(ctx).Model(&model.Tag{})
	if req.Search != "" {
		pattern := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(slug) LIKE ?", pattern, pattern)
	}
	if !admin || req.Status != "" {
		status := req.Status
		if !admin && status == "" {
			status = "active"
		}
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, common.NewInternal("failed to list tags", fmt.Errorf("count: %w", err))
	}

	var tags []model.Tag
	if err := query.Offset((page-1)*pageSize).Limit(pageSize).Order("id DESC").Find(&tags).Error; err != nil {
		return nil, common.NewInternal("failed to list tags", fmt.Errorf("find: %w", err))
	}

	items := make([]v1.TagItem, len(tags))
	for i, tag := range tags {
		item := toTagItem(tag)
		var count int64
		t.svc.DB.WithContext(ctx).Model(&model.ContentTag{}).Where("tag_id = ?", tag.ID).Count(&count)
		item.ContentCount = count
		items[i] = item
	}

	return &v1.ListTagsResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// List handles GET /api/v1/tags.
// List 处理 GET /api/v1/tags。
func (t *TagsController) List(ctx *gin.Context) {
	var req v1.ListTagsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid query parameters", err))
		return
	}
	actor := actorFromContext(ctx)
	resp, err := t.list(ctx.Request.Context(), &req, actor.isAdmin())
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// create creates a new tag.
// create 创建新标签。
func (t *TagsController) create(ctx context.Context, req *v1.CreateTagRequest) (*v1.CreateTagResponse, error) {
	tag := model.Tag{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Color:       req.Color,
		Status:      "active",
	}
	if err := t.svc.DB.WithContext(ctx).Create(&tag).Error; err != nil {
		return nil, mapTagCrudUniqueViolation(err, req.Slug)
	}
	return &v1.CreateTagResponse{ID: tag.ID}, nil
}

// Create handles POST /api/v1/tags (admin).
// Create 处理 POST /api/v1/tags（管理员）。
func (t *TagsController) Create(ctx *gin.Context) {
	var req v1.CreateTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := t.create(ctx.Request.Context(), &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// get returns a single tag by ID.
// get 根据 ID 返回单个标签。
func (t *TagsController) get(ctx context.Context, id int64) (*v1.TagDetailResponse, error) {
	var tag model.Tag
	if err := t.svc.DB.WithContext(ctx).First(&tag, id).Error; err != nil {
		return nil, common.NewNotFound("tag not found", err)
	}

	var contentCount int64
	t.svc.DB.WithContext(ctx).Model(&model.ContentTag{}).Where("tag_id = ?", tag.ID).Count(&contentCount)

	item := toTagItem(tag)
	item.ContentCount = contentCount

	return &v1.TagDetailResponse{TagItem: item}, nil
}

// Get handles GET /api/v1/tags/:id.
// Get 处理 GET /api/v1/tags/:id。
func (t *TagsController) Get(ctx *gin.Context) {
	id, ok := parseTagID(ctx)
	if !ok {
		return
	}
	resp, err := t.get(ctx.Request.Context(), id)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// update performs a PATCH update on a tag.
// update 对标签执行 PATCH 更新。
func (t *TagsController) update(ctx context.Context, id int64, req *v1.UpdateTagRequest) (*v1.TagDetailResponse, error) {
	var tag model.Tag
	if err := t.svc.DB.WithContext(ctx).First(&tag, id).Error; err != nil {
		return nil, common.NewNotFound("tag not found", err)
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Slug != nil {
		updates["slug"] = *req.Slug
	}
	if req.Description != nil {
		if *req.Description == "" {
			updates["description"] = nil
		} else {
			updates["description"] = *req.Description
		}
	}
	if req.Color != nil {
		if *req.Color == "" {
			updates["color"] = nil
		} else {
			updates["color"] = *req.Color
		}
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		return t.get(ctx, id)
	}

	if err := t.svc.DB.WithContext(ctx).Model(&tag).Updates(updates).Error; err != nil {
		return nil, mapTagCrudUniqueViolation(err, tag.Slug)
	}
	return t.get(ctx, id)
}

// Update handles PATCH /api/v1/tags/:id (admin).
// Update 处理 PATCH /api/v1/tags/:id（管理员）。
func (t *TagsController) Update(ctx *gin.Context) {
	id, ok := parseTagID(ctx)
	if !ok {
		return
	}
	var req v1.UpdateTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := t.update(ctx.Request.Context(), id, &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// del soft-deletes a tag after checking for existing content references.
// del 在检查现有内容引用后软删除标签。
func (t *TagsController) del(ctx context.Context, id int64) error {
	var tag model.Tag
	if err := t.svc.DB.WithContext(ctx).First(&tag, id).Error; err != nil {
		return common.NewNotFound("tag not found", err)
	}

	var refCount int64
	if err := t.svc.DB.WithContext(ctx).Model(&model.ContentTag{}).Where("tag_id = ?", id).Count(&refCount).Error; err != nil {
		return common.NewInternal("failed to check tag references", err)
	}
	if refCount > 0 {
		return common.NewConflict(
			fmt.Sprintf("tag is referenced by %d content item(s); remove references first", refCount),
			fmt.Errorf("tag %d has %d content references", id, refCount),
		)
	}

	if err := t.svc.DB.WithContext(ctx).Delete(&tag).Error; err != nil {
		return common.NewInternal("failed to delete tag", err)
	}
	return nil
}

// Delete handles DELETE /api/v1/tags/:id (admin).
// Delete 处理 DELETE /api/v1/tags/:id（管理员）。
func (t *TagsController) Delete(ctx *gin.Context) {
	id, ok := parseTagID(ctx)
	if !ok {
		return
	}
	if err := t.del(ctx.Request.Context(), id); err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, nil)
}

// actor holds optional caller info for admin detection on public routes.
// actor 保存可选调用者信息，用于在公开路由上检测管理员。
type actor struct {
	uid  int64
	role string
}

func (a actor) isAdmin() bool {
	return a.role == "admin"
}

// actorFromContext extracts optional actor info from the Gin context.
// Returns zero-value actor if no valid claims are present (anonymous).
// actorFromContext 从 Gin 上下文提取可选调用者信息。若无有效 claims 则返回零值（匿名）。
func actorFromContext(ctx *gin.Context) actor {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		return actor{}
	}
	return actor{uid: claims.UID, role: claims.Role}
}
