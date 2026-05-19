package tags

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// list queries tags with pagination and optional filters.
// list 查询标签列表（分页+可选筛选）。
func (t *TagsController) list(ctx context.Context, req *v1.ListTagsRequest, admin bool) (interface{}, error) {
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

	// Non-admin always sees active tags only; admin can filter by status.
	// 非管理员仅看到 active 标签；管理员可按状态筛选。
	if !admin {
		query = query.Where("status = ?", "active")
	} else if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, common.NewInternal("failed to list tags", fmt.Errorf("count: %w", err))
	}

	var tags []model.Tag
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&tags).Error; err != nil {
		return nil, common.NewInternal("failed to list tags", fmt.Errorf("find: %w", err))
	}

	// Batch-load content counts. / 批量加载内容数量。
	tagIDs := make([]int64, len(tags))
	for i, tag := range tags {
		tagIDs[i] = tag.ID
	}
	countMap := batchTagContentCounts(ctx, t, tagIDs)

	if !admin {
		items := make([]v1.TagItem, len(tags))
		for i, tag := range tags {
			item := toPublicTagItem(tag)
			item.ContentCount = countMap[tag.ID]
			items[i] = item
		}
		return &v1.ListTagsResponse{
			Items:    items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		}, nil
	}

	items := make([]v1.TagItem, len(tags))
	for i, tag := range tags {
		item := toTagItem(tag)
		item.ContentCount = countMap[tag.ID]
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
func (t *TagsController) get(ctx context.Context, id int64, admin bool) (*v1.TagDetailResponse, error) {
	var tag model.Tag
	if err := t.svc.DB.WithContext(ctx).First(&tag, id).Error; err != nil {
		return nil, common.NewNotFound("tag not found", err)
	}

	var contentCount int64
	if err := t.svc.DB.WithContext(ctx).Model(&model.ContentTag{}).Where("tag_id = ?", tag.ID).Count(&contentCount).Error; err != nil {
		return nil, common.NewInternal("failed to count tag content", err)
	}

	if !admin {
		item := toPublicTagItem(tag)
		item.ContentCount = contentCount
		return &v1.TagDetailResponse{TagItem: item}, nil
	}

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
	actor := actorFromContext(ctx)
	resp, err := t.get(ctx.Request.Context(), id, actor.isAdmin())
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
	newSlug := tag.Slug
	if req.Slug != nil {
		newSlug = *req.Slug
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
		return t.get(ctx, id, true)
	}

	if err := t.svc.DB.WithContext(ctx).Model(&tag).Updates(updates).Error; err != nil {
		return nil, mapTagCrudUniqueViolation(err, newSlug)
	}
	return t.get(ctx, id, true)
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

// batchTagContentCounts loads content counts for multiple tag IDs in one query.
// batchTagContentCounts 通过一次查询批量加载多个标签的内容数量。
func batchTagContentCounts(ctx context.Context, ctrl *TagsController, tagIDs []int64) map[int64]int64 {
	if len(tagIDs) == 0 {
		return nil
	}
	type countRow struct {
		TagID int64
		Count int64
	}
	var rows []countRow
	if err := ctrl.svc.DB.WithContext(ctx).Model(&model.ContentTag{}).
		Select("tag_id, COUNT(*) as count").
		Where("tag_id IN ?", tagIDs).
		Group("tag_id").
		Find(&rows).Error; err != nil {
		return nil
	}
	m := make(map[int64]int64, len(rows))
	for _, r := range rows {
		m[r.TagID] = r.Count
	}
	return m
}
