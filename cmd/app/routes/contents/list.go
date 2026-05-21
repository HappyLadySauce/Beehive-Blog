package contents

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/pagination"
)

// list queries contents with pagination and optional filters.
// list 查询内容列表（分页+可选筛选）。
func (c *ContentsController) list(ctx context.Context, req *v1.ListContentsRequest, admin bool) (interface{}, error) {
	page, pageSize := pagination.NormalizeOffset(req.Page, req.PageSize)

	query := c.svc.DB.WithContext(ctx).Model(&model.Content{})

	if !admin {
		query = query.Where("status = ? AND visibility = ?", "published", "public")
	} else {
		if req.Status != "" {
			query = query.Where("status = ?", req.Status)
		}
		if req.Visibility != "" {
			query = query.Where("visibility = ?", req.Visibility)
		}
	}

	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Slug != "" {
		query = query.Where("slug = ?", req.Slug)
	}
	if req.TagID > 0 {
		subQuery := c.svc.DB.WithContext(ctx).Model(&model.ContentTag{}).
			Select("content_id").Where("tag_id = ?", req.TagID)
		query = query.Where("id IN (?)", subQuery)
	}
	if req.Search != "" {
		pattern := "%" + req.Search + "%"
		query = query.Where("title ILIKE ? OR COALESCE(excerpt, '') ILIKE ?", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, common.NewInternal("failed to list contents", fmt.Errorf("count: %w", err))
	}

	var contents []model.Content
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&contents).Error; err != nil {
		return nil, common.NewInternal("failed to list contents", fmt.Errorf("find: %w", err))
	}

	// Batch-load author usernames. / 批量加载作者用户名。
	authorIDs := make([]int64, len(contents))
	contentIDs := make([]int64, len(contents))
	for i, ct := range contents {
		authorIDs[i] = ct.AuthorID
		contentIDs[i] = ct.ID
	}

	authorMap, err := batchLoadAuthorUsernames(ctx, c, authorIDs)
	if err != nil {
		return nil, common.NewInternal("failed to load author usernames", err)
	}
	tagMap, err := batchLoadContentTags(ctx, c, contentIDs)
	if err != nil {
		return nil, common.NewInternal("failed to load content tags", err)
	}

	if !admin {
		items := make([]v1.PublicContentItem, len(contents))
		for i, content := range contents {
			item := toPublicContentItem(content)
			item.AuthorUsername = authorMap[content.AuthorID]
			item.Tags = tagMap[content.ID]
			items[i] = item
		}
		return &v1.PublicListContentsResponse{
			Items:    items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		}, nil
	}

	items := make([]v1.ContentItem, len(contents))
	for i, content := range contents {
		item := toContentItem(content)
		item.AuthorUsername = authorMap[content.AuthorID]
		item.Tags = tagMap[content.ID]
		items[i] = item
	}

	return &v1.ListContentsResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// List handles GET /api/v1/contents.
// List 处理 GET /api/v1/contents。
//
//	@Summary		List contents
//	@Description	Paginated list of contents with optional filters. Without token: data is v1.PublicListContentsResponse (published public only). With admin Bearer: data is v1.ListContentsResponse (all statuses). 中文：无 token 时 data 为 PublicListContentsResponse（仅公开已发布）；管理员 Bearer 时 data 为 ListContentsResponse（含草稿等全部状态）。
//	@Tags			contents
//	@Produce		json
//	@Param			page		query		int		false	"Page number (default 1)"						default(1)
//	@Param			page_size	query		int		false	"Items per page (default 20, max 100)"			default(20)
//	@Param			type		query		string	false	"Filter by content type"						Enums(article, note, project, experience, reflection, portfolio)
//	@Param			status		query		string	false	"Filter by status (admin only)"				Enums(draft, review, published, archived)
//	@Param			visibility	query		string	false	"Filter by visibility (admin only)"			Enums(public, member, private)
//	@Param			tag_id		query		int		false	"Filter by tag ID"
//	@Param			search		query		string	false	"Search title or excerpt"
//	@Param			slug		query		string	false	"Exact slug match (combine with type for uniqueness)"
//	@Success		200			{object}	common.BaseResponse{data=v1.ListContentsResponse}	"Admin view; anonymous uses PublicListContentsResponse shape"
//	@Failure		400			{object}	common.BaseResponse
//	@Router			/api/v1/contents [get]
func (c *ContentsController) List(ctx *gin.Context) {
	var req v1.ListContentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid query parameters", err))
		return
	}
	claims := middleware.GetClaims(ctx)
	admin := claims != nil && claims.Role == "admin"
	resp, err := c.list(ctx.Request.Context(), &req, admin)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// batchLoadAuthorUsernames loads usernames for a set of author IDs in one query.
// batchLoadAuthorUsernames 一次查询批量加载作者用户名。
func batchLoadAuthorUsernames(ctx context.Context, ctrl *ContentsController, ids []int64) (map[int64]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	uniqueIDs := uniqueInt64(ids)
	var users []model.User
	if err := ctrl.svc.DB.WithContext(ctx).Select("id, username").Where("id IN ?", uniqueIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	m := make(map[int64]string, len(users))
	for _, u := range users {
		m[u.ID] = u.Username
	}
	return m, nil
}

// fetchTagsByIDs loads tag rows (non-soft-deleted via GORM scope).
// fetchTagsByIDs 加载标签行（GORM 默认 scope 排除软删）。
func fetchTagsByIDs(ctx context.Context, ctrl *ContentsController, tagIDs []int64) ([]model.Tag, error) {
	if len(tagIDs) == 0 {
		return nil, nil
	}
	var tags []model.Tag
	if err := ctrl.svc.DB.WithContext(ctx).Where("id IN ?", uniqueInt64(tagIDs)).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func tagsToItems(tags []model.Tag) []v1.TagItem {
	items := make([]v1.TagItem, len(tags))
	for i, t := range tags {
		items[i] = v1.TagItem{
			ID:    t.ID,
			Name:  t.Name,
			Slug:  t.Slug,
			Color: t.Color,
		}
	}
	return items
}

// batchLoadContentTags loads tags for multiple content IDs in two queries.
// batchLoadContentTags 通过两次查询批量加载多个内容的标签。
func batchLoadContentTags(ctx context.Context, ctrl *ContentsController, contentIDs []int64) (map[int64][]v1.TagItem, error) {
	if len(contentIDs) == 0 {
		return nil, nil
	}
	// Load all junction rows. / 加载所有联结行。
	var cts []model.ContentTag
	if err := ctrl.svc.DB.WithContext(ctx).Where("content_id IN ?", contentIDs).Find(&cts).Error; err != nil {
		return nil, err
	}
	if len(cts) == 0 {
		return nil, nil
	}

	// Collect unique tag IDs. / 收集唯一标签 ID。
	tagIDs := make([]int64, len(cts))
	for i, ct := range cts {
		tagIDs[i] = ct.TagID
	}

	tags, err := fetchTagsByIDs(ctx, ctrl, tagIDs)
	if err != nil {
		return nil, err
	}
	tagItemMap := make(map[int64]v1.TagItem, len(tags))
	for _, t := range tags {
		tagItemMap[t.ID] = v1.TagItem{
			ID:    t.ID,
			Name:  t.Name,
			Slug:  t.Slug,
			Color: t.Color,
		}
	}

	// Assemble content → tags mapping. / 组装内容→标签映射。
	result := make(map[int64][]v1.TagItem, len(contentIDs))
	for _, ct := range cts {
		if item, ok := tagItemMap[ct.TagID]; ok {
			result[ct.ContentID] = append(result[ct.ContentID], item)
		}
	}
	return result, nil
}

// uniqueInt64 deduplicates an int64 slice while preserving order.
// uniqueInt64 对 int64 切片去重，保持顺序。
func uniqueInt64(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// loadContentTags fetches tags for a single content ID.
// loadContentTags 获取单个内容的标签列表。
func loadContentTags(ctx context.Context, ctrl *ContentsController, contentID int64) ([]v1.TagItem, error) {
	var cts []model.ContentTag
	if err := ctrl.svc.DB.WithContext(ctx).Where("content_id = ?", contentID).Find(&cts).Error; err != nil {
		return nil, err
	}
	if len(cts) == 0 {
		return []v1.TagItem{}, nil
	}
	tagIDs := make([]int64, len(cts))
	for i, ct := range cts {
		tagIDs[i] = ct.TagID
	}
	tags, err := fetchTagsByIDs(ctx, ctrl, tagIDs)
	if err != nil {
		return nil, err
	}
	return tagsToItems(tags), nil
}
