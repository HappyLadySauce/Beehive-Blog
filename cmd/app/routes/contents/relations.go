package contents

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// getRelations lists outgoing relations for a content item.
// getRelations 列出某个内容的出向关系。
func (c *ContentsController) getRelations(ctx context.Context, contentID int64, admin bool) (*v1.ListContentRelationsResponse, error) {
	existQuery := c.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id = ?", contentID)
	if !admin {
		existQuery = existQuery.Where("status = ? AND visibility = ?", "published", "public")
	}
	var count int64
	if err := existQuery.Count(&count).Error; err != nil || count == 0 {
		return nil, common.NewNotFound("content not found", fmt.Errorf("content %d not found", contentID))
	}

	var relations []model.ContentRelation
	if err := c.svc.DB.WithContext(ctx).
		Where("source_content_id = ?", contentID).
		Order("sort_order ASC, id ASC").
		Find(&relations).Error; err != nil {
		return nil, common.NewInternal("failed to list relations", err)
	}

	// Batch-load target content summaries. / 批量加载目标内容摘要。
	targetIDs := make([]int64, len(relations))
	for i, rel := range relations {
		targetIDs[i] = rel.TargetContentID
	}
	targetMap := batchLoadTargets(ctx, c, uniqueInt64(targetIDs), admin)

	items := make([]v1.ContentRelationItem, len(relations))
	for i, rel := range relations {
		item := v1.ContentRelationItem{
			ID:              rel.ID,
			SourceContentID: rel.SourceContentID,
			TargetContentID: rel.TargetContentID,
			RelationType:    rel.RelationType,
			Label:           rel.Label,
			SortOrder:       rel.SortOrder,
			CreatedAt:       rel.CreatedAt,
		}
		if tgt, ok := targetMap[rel.TargetContentID]; ok {
			item.TargetTitle = tgt.Title
			item.TargetType = tgt.Type
			item.TargetSlug = tgt.Slug
		}
		items[i] = item
	}

	return &v1.ListContentRelationsResponse{Items: items}, nil
}

// GetRelations handles GET /api/v1/contents/:id/relations.
// GetRelations 处理 GET /api/v1/contents/:id/relations。
func (c *ContentsController) GetRelations(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	claims := middleware.GetClaims(ctx)
	admin := claims != nil && claims.Role == "admin"
	resp, err := c.getRelations(ctx.Request.Context(), id, admin)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// addRelation adds a directed relation from content to a target.
// addRelation 添加从内容到目标的有向关系。
func (c *ContentsController) addRelation(ctx context.Context, contentID int64, req *v1.AddRelationRequest) (*v1.AddRelationResponse, error) {
	var count int64
	if err := c.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id = ?", contentID).Count(&count).Error; err != nil || count == 0 {
		return nil, common.NewNotFound("content not found", fmt.Errorf("content %d not found", contentID))
	}
	if err := c.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id = ?", req.TargetContentID).Count(&count).Error; err != nil || count == 0 {
		return nil, common.NewNotFound("target content not found", fmt.Errorf("target %d not found", req.TargetContentID))
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	relation := model.ContentRelation{
		SourceContentID: contentID,
		TargetContentID: req.TargetContentID,
		RelationType:    req.RelationType,
		Label:           req.Label,
		SortOrder:       sortOrder,
	}

	if err := c.svc.DB.WithContext(ctx).Create(&relation).Error; err != nil {
		return nil, common.NewInternal("failed to add relation", err)
	}

	item := v1.ContentRelationItem{
		ID:              relation.ID,
		SourceContentID: relation.SourceContentID,
		TargetContentID: relation.TargetContentID,
		RelationType:    relation.RelationType,
		Label:           relation.Label,
		SortOrder:       relation.SortOrder,
		CreatedAt:       relation.CreatedAt,
	}
	var target model.Content
	if err := c.svc.DB.WithContext(ctx).Select("title, type, slug").First(&target, req.TargetContentID).Error; err == nil {
		item.TargetTitle = target.Title
		item.TargetType = target.Type
		item.TargetSlug = target.Slug
	}

	return &v1.AddRelationResponse{ContentRelationItem: item}, nil
}

// AddRelation handles POST /api/v1/contents/:id/relations (admin).
// AddRelation 处理 POST /api/v1/contents/:id/relations（管理员）。
func (c *ContentsController) AddRelation(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	var req v1.AddRelationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := c.addRelation(ctx.Request.Context(), id, &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// removeRelation deletes a relation, verifying it belongs to the given content.
// removeRelation 删除关系，验证其属于指定内容。
func (c *ContentsController) removeRelation(ctx context.Context, contentID int64, relationID int64) error {
	var relation model.ContentRelation
	if err := c.svc.DB.WithContext(ctx).Where("id = ? AND source_content_id = ?", relationID, contentID).First(&relation).Error; err != nil {
		return common.NewNotFound("relation not found", err)
	}
	if err := c.svc.DB.WithContext(ctx).Delete(&relation).Error; err != nil {
		return common.NewInternal("failed to remove relation", err)
	}
	return nil
}

// RemoveRelation handles DELETE /api/v1/contents/:id/relations/:relationId (admin).
// RemoveRelation 处理 DELETE /api/v1/contents/:id/relations/:relationId（管理员）。
func (c *ContentsController) RemoveRelation(ctx *gin.Context) {
	contentID, ok := parseContentID(ctx)
	if !ok {
		return
	}
	relationID, ok := parseRelationID(ctx)
	if !ok {
		return
	}
	if err := c.removeRelation(ctx.Request.Context(), contentID, relationID); err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, nil)
}

// batchLoadTargets loads title/type/slug for target content IDs in one query.
// batchLoadTargets 一次查询批量加载目标内容的标题/类型/slug。
func batchLoadTargets(ctx context.Context, ctrl *ContentsController, ids []int64, admin bool) map[int64]struct {
	Title string
	Type  string
	Slug  string
} {
	if len(ids) == 0 {
		return nil
	}
	query := ctrl.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id IN ?", ids)
	if !admin {
		query = query.Where("status = ? AND visibility = ?", "published", "public")
	}
	type row struct {
		ID    int64
		Title string
		Type  string
		Slug  string
	}
	var rows []row
	if err := query.Select("id, title, type, slug").Find(&rows).Error; err != nil {
		return nil
	}
	m := make(map[int64]struct {
		Title string
		Type  string
		Slug  string
	}, len(rows))
	for _, r := range rows {
		m[r.ID] = struct {
			Title string
			Type  string
			Slug  string
		}{Title: r.Title, Type: r.Type, Slug: r.Slug}
	}
	return m
}
