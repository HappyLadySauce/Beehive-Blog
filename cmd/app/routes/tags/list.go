package tags

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/pagination"
)

// list queries tags with pagination and optional filters.
// list 查询标签列表（分页+可选筛选）。
func (t *TagsController) list(ctx context.Context, req *v1.ListTagsRequest) (*v1.ListTagsResponse, error) {
	page, pageSize := pagination.NormalizeOffset(req.Page, req.PageSize)

	query := t.svc.DB.WithContext(ctx).Model(&model.Tag{})
	if req.Search != "" {
		pattern := "%" + req.Search + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ?", pattern, pattern)
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
	countMap, err := batchTagContentCounts(ctx, t, tagIDs)
	if err != nil {
		return nil, common.NewInternal("failed to load tag content counts", err)
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
//
//	@Summary		List tags
//	@Description	Paginated list of non-deleted tags. Optional admin Bearer for management UIs. 中文：分页返回未软删标签；可选管理员 Bearer 用于管理端。
//	@Tags			tags
//	@Produce		json
//	@Param			page		query		int		false	"Page number (default 1)"				default(1)
//	@Param			page_size	query		int		false	"Items per page (default 20, max 100)"	default(20)
//	@Param			search		query		string	false	"Search name or slug"
//	@Success		200			{object}	common.BaseResponse{data=v1.ListTagsResponse}
//	@Failure		400			{object}	common.BaseResponse
//	@Router			/api/v1/tags [get]
func (t *TagsController) List(ctx *gin.Context) {
	var req v1.ListTagsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid query parameters", err))
		return
	}
	resp, err := t.list(ctx.Request.Context(), &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// batchTagContentCounts loads content counts for multiple tag IDs in one query.
// batchTagContentCounts 通过一次查询批量加载多个标签的内容数量。
func batchTagContentCounts(ctx context.Context, ctrl *TagsController, tagIDs []int64) (map[int64]int64, error) {
	if len(tagIDs) == 0 {
		return nil, nil
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
		return nil, err
	}
	m := make(map[int64]int64, len(rows))
	for _, r := range rows {
		m[r.TagID] = r.Count
	}
	return m, nil
}
