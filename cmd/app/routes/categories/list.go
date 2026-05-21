package categories

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/pagination"
)

// list queries categories with pagination and optional filters.
// list 查询分类列表（分页+可选筛选）。
func (c *CategoriesController) list(ctx context.Context, req *v1.ListCategoriesRequest) (*v1.ListCategoriesResponse, error) {
	page, pageSize := pagination.NormalizeOffset(req.Page, req.PageSize)

	query := c.svc.DB.WithContext(ctx).Model(&model.Category{})
	if req.Search != "" {
		pattern := "%" + req.Search + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ?", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, common.NewInternal("failed to list categories", fmt.Errorf("count: %w", err))
	}

	var cats []model.Category
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("sort_order ASC, id DESC").Find(&cats).Error; err != nil {
		return nil, common.NewInternal("failed to list categories", fmt.Errorf("find: %w", err))
	}

	// Batch-load content counts. / 批量加载内容数量。
	catIDs := make([]int64, len(cats))
	for i, cat := range cats {
		catIDs[i] = cat.ID
	}
	countMap, err := batchCategoryContentCounts(ctx, c, catIDs)
	if err != nil {
		return nil, common.NewInternal("failed to load category content counts", err)
	}

	items := make([]v1.CategoryItem, len(cats))
	for i, cat := range cats {
		item := toCategoryItem(cat)
		item.ContentCount = countMap[cat.ID]
		items[i] = item
	}

	return &v1.ListCategoriesResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// List handles GET /api/v1/categories.
// List 处理 GET /api/v1/categories。
//
//	@Summary		List categories
//	@Description	Paginated list of non-deleted categories. 中文：分页返回未软删分类。
//	@Tags			categories
//	@Produce		json
//	@Param			page		query		int		false	"Page number (default 1)"				default(1)
//	@Param			page_size	query		int		false	"Items per page (default 20, max 100)"	default(20)
//	@Param			search		query		string	false	"Search name or slug"
//	@Success		200			{object}	common.BaseResponse{data=v1.ListCategoriesResponse}
//	@Failure		400			{object}	common.BaseResponse
//	@Router			/api/v1/categories [get]
func (c *CategoriesController) List(ctx *gin.Context) {
	var req v1.ListCategoriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid query parameters", err))
		return
	}
	resp, err := c.list(ctx.Request.Context(), &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// batchCategoryContentCounts loads content counts for multiple category IDs in one query.
// batchCategoryContentCounts 通过一次查询批量加载多个分类的内容数量。
func batchCategoryContentCounts(ctx context.Context, ctrl *CategoriesController, catIDs []int64) (map[int64]int64, error) {
	if len(catIDs) == 0 {
		return nil, nil
	}
	type countRow struct {
		CategoryID int64
		Count      int64
	}
	var rows []countRow
	if err := ctrl.svc.DB.WithContext(ctx).Model(&model.ContentCategory{}).
		Select("category_id, COUNT(*) as count").
		Where("category_id IN ?", catIDs).
		Group("category_id").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[int64]int64, len(rows))
	for _, r := range rows {
		m[r.CategoryID] = r.Count
	}
	return m, nil
}
