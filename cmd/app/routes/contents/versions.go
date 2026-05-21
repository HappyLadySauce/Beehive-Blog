package contents

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

const (
	versionSnapshotManual = "manual"
	versionSnapshotAuto   = "auto"
	autoVersionName       = "自动保存"
)

// listVersions lists version snapshots for a content item.
// listVersions 列出某个内容的版本快照。
func (c *ContentsController) listVersions(ctx context.Context, contentID int64) (*v1.ListVersionsResponse, error) {
	var count int64
	if err := c.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id = ?", contentID).Count(&count).Error; err != nil {
		return nil, common.NewInternal("failed to check content", err)
	}
	if count == 0 {
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
//
//	@Summary		List content versions
//	@Description	Lists all version snapshots for a content item. Admin only. 中文：列出内容的全部版本快照（仅管理员）。
//	@Tags			contents
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Content ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.ListVersionsResponse}
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Router			/api/v1/contents/{id}/versions [get]
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

// createVersion creates or updates a version snapshot from the current content state inside a transaction.
// Manual snapshots are append-only; auto snapshots are overwritten per content item.
// createVersion 在事务中根据当前内容状态创建或更新版本快照。
// 手动快照追加保存；自动快照按内容覆盖保存。
func (c *ContentsController) createVersion(ctx context.Context, contentID int64, createdBy int64, req *v1.CreateVersionRequest) (*v1.CreateVersionResponse, error) {
	snapshotType := versionSnapshotManual
	if req.SnapshotType != nil && *req.SnapshotType != "" {
		snapshotType = *req.SnapshotType
	}
	name := ""
	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
	}
	if snapshotType == versionSnapshotManual && name == "" {
		return nil, common.NewBadRequest("version name is required", fmt.Errorf("manual version name is empty"))
	}
	if snapshotType == versionSnapshotAuto {
		name = autoVersionName
	}

	tx := c.svc.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, common.NewInternal("failed to begin transaction", tx.Error)
	}

	var content model.Content
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&content, contentID).Error; err != nil {
		tx.Rollback()
		return nil, mapFirstError(err, "content not found", "failed to fetch content")
	}

	if snapshotType == versionSnapshotAuto {
		var existing model.ContentVersion
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("content_id = ? AND snapshot_type = ?", contentID, versionSnapshotAuto).
			First(&existing).Error
		if err == nil {
			updates := map[string]interface{}{
				"name":           name,
				"title":          content.Title,
				"body":           content.Body,
				"excerpt":        content.Excerpt,
				"change_summary": req.ChangeSummary,
				"created_by":     createdBy,
			}
			if err := tx.Model(&existing).Updates(updates).Error; err != nil {
				tx.Rollback()
				return nil, common.NewInternal("failed to update auto version", err)
			}
			if err := tx.Commit().Error; err != nil {
				return nil, common.NewInternal("failed to commit version creation", err)
			}
			existing.Name = name
			existing.Title = content.Title
			existing.Body = content.Body
			existing.Excerpt = content.Excerpt
			existing.ChangeSummary = req.ChangeSummary
			existing.CreatedBy = createdBy
			item := toVersionItem(existing)
			return &v1.CreateVersionResponse{VersionItem: item}, nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return nil, common.NewInternal("failed to fetch auto version", err)
		}
	}

	var maxVersion int
	if err := tx.Model(&model.ContentVersion{}).
		Select("COALESCE(MAX(version_number), 0)").
		Where("content_id = ?", contentID).
		Scan(&maxVersion).Error; err != nil {
		tx.Rollback()
		return nil, common.NewInternal("failed to read version number", err)
	}

	version := model.ContentVersion{
		ContentID:     contentID,
		VersionNumber: maxVersion + 1,
		SnapshotType:  snapshotType,
		Name:          name,
		Title:         content.Title,
		Body:          content.Body,
		Excerpt:       content.Excerpt,
		ChangeSummary: req.ChangeSummary,
		CreatedBy:     createdBy,
	}

	if err := tx.Create(&version).Error; err != nil {
		tx.Rollback()
		return nil, mapVersionUniqueViolation(err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, common.NewInternal("failed to commit version creation", err)
	}

	item := toVersionItem(version)
	return &v1.CreateVersionResponse{VersionItem: item}, nil
}

// CreateVersion handles POST /api/v1/contents/:id/versions (admin).
// CreateVersion 处理 POST /api/v1/contents/:id/versions（管理员）。
//
//	@Summary		Create content version
//	@Description	Creates a version snapshot from current content state. Admin only. 中文：从当前内容状态创建版本快照（仅管理员）。
//	@Tags			contents
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Content ID"
//	@Param			body	body		v1.CreateVersionRequest	true	"Version creation request (optional change_summary)"
//	@Success		200		{object}	common.BaseResponse{data=v1.CreateVersionResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Router			/api/v1/contents/{id}/versions [post]
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

// deleteVersion removes one version snapshot for a content item.
// deleteVersion 删除某篇内容的一个版本快照。
func (c *ContentsController) deleteVersion(ctx context.Context, contentID int64, versionNumber int) error {
	result := c.svc.DB.WithContext(ctx).
		Where("content_id = ? AND version_number = ?", contentID, versionNumber).
		Delete(&model.ContentVersion{})
	if result.Error != nil {
		return common.NewInternal("failed to delete version", result.Error)
	}
	if result.RowsAffected == 0 {
		return common.NewNotFound("version not found", fmt.Errorf("content %d version %d not found", contentID, versionNumber))
	}
	return nil
}

// DeleteVersion handles DELETE /api/v1/contents/:id/versions/:versionNumber (admin).
// DeleteVersion 处理 DELETE /api/v1/contents/:id/versions/:versionNumber（管理员）。
//
//	@Summary		Delete content version
//	@Description	Deletes one version snapshot for a content item. Admin only. 中文：删除某篇内容的一个版本快照（仅管理员）。
//	@Tags			contents
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id				path	int	true	"Content ID"
//	@Param			versionNumber	path	int	true	"Version number to delete"
//	@Success		200				{object}	common.BaseResponse{data=map[string]interface{}}
//	@Failure		400				{object}	common.BaseResponse
//	@Failure		401				{object}	common.BaseResponse
//	@Failure		403				{object}	common.BaseResponse
//	@Failure		404				{object}	common.BaseResponse
//	@Router			/api/v1/contents/{id}/versions/{versionNumber} [delete]
func (c *ContentsController) DeleteVersion(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	vn, ok := parseVersionNumber(ctx)
	if !ok {
		return
	}
	if err := c.deleteVersion(ctx.Request.Context(), id, vn); err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, gin.H{})
}

// restoreVersion restores a version snapshot back to the content row inside a transaction.
// An auto snapshot is created before restore to prevent data loss.
// restoreVersion 在事务中将版本快照写回内容行。回滚前自动创建快照以防止数据丢失。
func (c *ContentsController) restoreVersion(ctx context.Context, contentID int64, versionNumber int, createdBy int64, req *v1.RestoreVersionRequest) (*v1.ContentDetailResponse, error) {
	tx := c.svc.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, common.NewInternal("failed to begin transaction", tx.Error)
	}

	var content model.Content
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&content, contentID).Error; err != nil {
		tx.Rollback()
		return nil, mapFirstError(err, "content not found", "failed to fetch content")
	}

	var version model.ContentVersion
	if err := tx.Where("content_id = ? AND version_number = ?", contentID, versionNumber).First(&version).Error; err != nil {
		tx.Rollback()
		return nil, mapFirstError(err, "version not found", "failed to fetch version")
	}

	// Auto snapshot the current state before restoring.
	// 回滚前对当前状态自动创建快照。
	var maxVersion int
	if err := tx.Model(&model.ContentVersion{}).
		Select("COALESCE(MAX(version_number), 0)").
		Where("content_id = ?", contentID).
		Scan(&maxVersion).Error; err != nil {
		tx.Rollback()
		return nil, common.NewInternal("failed to read version number", err)
	}

	snapshotSummary := fmt.Sprintf("Restored from v%d; auto snapshot before restore", versionNumber)
	if req.ChangeSummary != nil && *req.ChangeSummary != "" {
		snapshotSummary = *req.ChangeSummary
	}
	snapshot := model.ContentVersion{
		ContentID:     contentID,
		VersionNumber: maxVersion + 1,
		SnapshotType:  versionSnapshotManual,
		Name:          "恢复前备份",
		Title:         content.Title,
		Body:          content.Body,
		Excerpt:       content.Excerpt,
		ChangeSummary: &snapshotSummary,
		CreatedBy:     createdBy,
	}
	if err := tx.Create(&snapshot).Error; err != nil {
		tx.Rollback()
		return nil, mapVersionUniqueViolation(err)
	}

	// Write version snapshot fields back to the content row.
	// 将快照字段写回内容行。
	updates := map[string]interface{}{
		"title":   version.Title,
		"body":    version.Body,
		"excerpt": version.Excerpt,
	}
	wc := computeWordCount(version.Body)
	updates["word_count"] = wc
	updates["reading_time_minutes"] = computeReadingTime(wc)

	if err := tx.Model(&content).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, common.NewInternal("failed to restore content from version", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, common.NewInternal("failed to commit version restore", err)
	}

	resp, err := c.get(ctx, contentID, true)
	if err != nil {
		return nil, err
	}
	return resp.(*v1.ContentDetailResponse), nil
}

// RestoreVersion handles POST /api/v1/contents/:id/versions/:versionNumber/restore (admin).
// RestoreVersion 处理 POST /api/v1/contents/:id/versions/:versionNumber/restore（管理员）。
//
//	@Summary		Restore content version
//	@Description	Restores a version snapshot back to the current content. An auto snapshot is created before restore. Admin only. 中文：将版本快照恢复到当前内容。回滚前自动创建快照（仅管理员）。
//	@Tags			contents
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id				path		int							true	"Content ID"
//	@Param			versionNumber	path		int							true	"Version number to restore"
//	@Param			body			body		v1.RestoreVersionRequest	true	"Optional restore metadata"
//	@Success		200				{object}	common.BaseResponse{data=v1.ContentDetailResponse}
//	@Failure		400				{object}	common.BaseResponse
//	@Failure		401				{object}	common.BaseResponse
//	@Failure		403				{object}	common.BaseResponse
//	@Failure		404				{object}	common.BaseResponse
//	@Router			/api/v1/contents/{id}/versions/{versionNumber}/restore [post]
func (c *ContentsController) RestoreVersion(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	vn, ok := parseVersionNumber(ctx)
	if !ok {
		return
	}
	var req v1.RestoreVersionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	claims := middleware.GetClaims(ctx)
	createdBy := int64(0)
	if claims != nil {
		createdBy = claims.UID
	}
	resp, err := c.restoreVersion(ctx.Request.Context(), id, vn, createdBy, &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
