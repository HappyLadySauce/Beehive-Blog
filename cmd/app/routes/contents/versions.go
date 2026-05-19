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

// listVersions lists version snapshots for a content item.
// listVersions 列出某个内容的版本快照。
func (c *ContentsController) listVersions(ctx context.Context, contentID int64) (*v1.ListVersionsResponse, error) {
	var count int64
	if err := c.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id = ?", contentID).Count(&count).Error; err != nil || count == 0 {
		return nil, common.NewNotFound("content not found", fmt.Errorf("content %d not found", contentID))
	}

	var versions []model.ContentVersion
	if err := c.svc.DB.WithContext(ctx).
		Where("content_id = ?", contentID).
		Order("version_number DESC").
		Find(&versions).Error; err != nil {
		return nil, common.NewInternal("failed to list versions", err)
	}

	items := make([]v1.VersionItem, len(versions))
	for i, v := range versions {
		items[i] = toVersionItem(v)
	}

	return &v1.ListVersionsResponse{Items: items}, nil
}

// ListVersions handles GET /api/v1/contents/:id/versions (admin).
// ListVersions 处理 GET /api/v1/contents/:id/versions（管理员）。
func (c *ContentsController) ListVersions(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	resp, err := c.listVersions(ctx.Request.Context(), id)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// createVersion creates a version snapshot from the current content state.
// createVersion 根据当前内容状态创建版本快照。
func (c *ContentsController) createVersion(ctx context.Context, contentID int64, createdBy int64, req *v1.CreateVersionRequest) (*v1.CreateVersionResponse, error) {
	var content model.Content
	if err := c.svc.DB.WithContext(ctx).First(&content, contentID).Error; err != nil {
		return nil, common.NewNotFound("content not found", err)
	}

	var maxVersion int
	c.svc.DB.WithContext(ctx).Model(&model.ContentVersion{}).
		Select("COALESCE(MAX(version_number), 0)").
		Where("content_id = ?", contentID).
		Scan(&maxVersion)

	version := model.ContentVersion{
		ContentID:     contentID,
		VersionNumber: maxVersion + 1,
		Title:         content.Title,
		Body:          content.Body,
		Excerpt:       content.Excerpt,
		ChangeSummary: req.ChangeSummary,
		CreatedBy:     createdBy,
	}

	if err := c.svc.DB.WithContext(ctx).Create(&version).Error; err != nil {
		return nil, common.NewInternal("failed to create version", err)
	}

	item := toVersionItem(version)
	return &v1.CreateVersionResponse{VersionItem: item}, nil
}

// CreateVersion handles POST /api/v1/contents/:id/versions (admin).
// CreateVersion 处理 POST /api/v1/contents/:id/versions（管理员）。
func (c *ContentsController) CreateVersion(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	var req v1.CreateVersionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	claims := middleware.GetClaims(ctx)
	createdBy := int64(0)
	if claims != nil {
		createdBy = claims.UID
	}
	resp, err := c.createVersion(ctx.Request.Context(), id, createdBy, &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
