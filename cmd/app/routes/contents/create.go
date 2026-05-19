package contents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// create creates a new content row inside a transaction.
// create 在事务中创建新内容行。
func (c *ContentsController) create(ctx context.Context, authorID int64, req *v1.CreateContentRequest) (*v1.CreateContentResponse, error) {
	status := "draft"
	if req.Status != nil && *req.Status != "" {
		status = *req.Status
	}
	visibility := "public"
	if req.Visibility != nil && *req.Visibility != "" {
		visibility = *req.Visibility
	}
	aiAccess := "allowed"
	if req.AIAccess != nil && *req.AIAccess != "" {
		aiAccess = *req.AIAccess
	}

	metadata := json.RawMessage("{}")
	if len(req.Metadata) > 0 {
		metadata = req.Metadata
	}

	wc := computeWordCount(req.Body)
	rt := computeReadingTime(wc)
	if req.WordCount != nil {
		wc = *req.WordCount
	}
	if req.ReadingTimeMinutes != nil {
		rt = *req.ReadingTimeMinutes
	}

	var publishedAt *time.Time
	if status == "published" {
		now := time.Now()
		publishedAt = &now
	}

	tx := c.svc.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, common.NewInternal("failed to begin transaction", tx.Error)
	}

	// Validate slug uniqueness inside the transaction.
	// 在事务内校验 slug 唯一性。
	var existing int64
	if err := tx.Model(&model.Content{}).
		Where("type = ? AND slug = ? AND deleted_at IS NULL", req.Type, req.Slug).
		Count(&existing).Error; err != nil {
		tx.Rollback()
		return nil, common.NewInternal("failed to validate slug", err)
	}
	if existing > 0 {
		tx.Rollback()
		return nil, common.NewConflict(
			fmt.Sprintf("slug %q is already taken for type %q", req.Slug, req.Type),
			fmt.Errorf("slug conflict: %s/%s", req.Type, req.Slug),
		)
	}

	content := model.Content{
		Type:               req.Type,
		Title:              req.Title,
		Slug:               req.Slug,
		Excerpt:            req.Excerpt,
		Body:               req.Body,
		CoverAttachmentID:  req.CoverAttachmentID,
		AuthorID:           authorID,
		Status:             status,
		Visibility:         visibility,
		AIAccess:           aiAccess,
		PublishedAt:        publishedAt,
		WordCount:          wc,
		ReadingTimeMinutes: rt,
		Metadata:           metadata,
	}

	if err := tx.Create(&content).Error; err != nil {
		tx.Rollback()
		return nil, mapContentCrudUniqueViolation(err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, common.NewInternal("failed to commit content creation", err)
	}
	return &v1.CreateContentResponse{ID: content.ID}, nil
}

// Create handles POST /api/v1/contents (admin).
// Create 处理 POST /api/v1/contents（管理员）。
func (c *ContentsController) Create(ctx *gin.Context) {
	var req v1.CreateContentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}

	claims := middleware.GetClaims(ctx)
	if claims == nil {
		common.Fail(ctx, common.NewUnauthorized("authentication required", fmt.Errorf("no claims")))
		return
	}

	resp, err := c.create(ctx.Request.Context(), claims.UID, &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}

	common.Success(ctx, resp)
}
