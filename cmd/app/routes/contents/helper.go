package contents

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// mapFirstError maps gorm.ErrRecordNotFound to 404 and other errors to 500.
// mapFirstError 将 gorm.ErrRecordNotFound 映射为 404，其余错误映射为 500。
func mapFirstError(err error, notFoundMsg, internalMsg string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return common.NewNotFound(notFoundMsg, err)
	}
	return common.NewInternal(internalMsg, err)
}

// parseContentID extracts the :id path parameter as int64.
// parseContentID 将 :id 路径参数提取为 int64。
func parseContentID(ctx *gin.Context) (int64, bool) {
	raw := ctx.Param("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		common.Fail(ctx, common.NewBadRequest("invalid content id", fmt.Errorf("parse: %w", err)))
		return 0, false
	}
	return id, true
}

// parseVersionNumber extracts the :versionNumber path parameter as positive int.
// parseVersionNumber 将 :versionNumber 路径参数提取为正整数。
func parseVersionNumber(ctx *gin.Context) (int, bool) {
	raw := ctx.Param("versionNumber")
	vn, err := strconv.Atoi(raw)
	if err != nil || vn < 1 {
		common.Fail(ctx, common.NewBadRequest("versionNumber must be a positive integer", fmt.Errorf("parse: %w", err)))
		return 0, false
	}
	return vn, true
}

// parseRelationID extracts the :relationId path parameter as int64.
// parseRelationID 将 :relationId 路径参数提取为 int64。
func parseRelationID(ctx *gin.Context) (int64, bool) {
	raw := ctx.Param("relationId")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		common.Fail(ctx, common.NewBadRequest("invalid relation id", fmt.Errorf("parse: %w", err)))
		return 0, false
	}
	return id, true
}

// toContentItem converts a model.Content to its admin API response item.
// toContentItem 将 model.Content 转换为管理员 API 响应项。
func toContentItem(c model.Content) v1.ContentItem {
	return v1.ContentItem{
		ID:                 c.ID,
		Type:               c.Type,
		Title:              c.Title,
		Slug:               c.Slug,
		Excerpt:            c.Excerpt,
		CoverAttachmentID:  c.CoverAttachmentID,
		AuthorID:           c.AuthorID,
		Status:             c.Status,
		Visibility:         c.Visibility,
		AIAccess:           c.AIAccess,
		PublishedAt:        c.PublishedAt,
		WordCount:          c.WordCount,
		ReadingTimeMinutes: c.ReadingTimeMinutes,
		Metadata:           c.Metadata,
		ViewCount:          c.ViewCount,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
	}
}

// toPublicContentItem converts a model.Content to its public API response item.
// toPublicContentItem 将 model.Content 转换为公开 API 响应项（不含 metadata）。
func toPublicContentItem(c model.Content) v1.PublicContentItem {
	return v1.PublicContentItem{
		ID:                 c.ID,
		Type:               c.Type,
		Title:              c.Title,
		Slug:               c.Slug,
		Excerpt:            c.Excerpt,
		CoverAttachmentID:  c.CoverAttachmentID,
		AuthorID:           c.AuthorID,
		PublishedAt:        c.PublishedAt,
		WordCount:          c.WordCount,
		ReadingTimeMinutes: c.ReadingTimeMinutes,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
	}
}

// toVersionItem converts a model.ContentVersion to its API response item.
// toVersionItem 将 model.ContentVersion 转换为 API 响应项。
func toVersionItem(v model.ContentVersion) v1.VersionItem {
	return v1.VersionItem{
		ID:            v.ID,
		ContentID:     v.ContentID,
		VersionNumber: v.VersionNumber,
		SnapshotType:  v.SnapshotType,
		Name:          v.Name,
		Title:         v.Title,
		Body:          v.Body,
		Excerpt:       v.Excerpt,
		ChangeSummary: v.ChangeSummary,
		CreatedBy:     v.CreatedBy,
		CreatedAt:     v.CreatedAt,
	}
}

// computeWordCount returns the word count of a body string.
// computeWordCount 返回正文字符串的字数。
func computeWordCount(body *string) int {
	if body == nil || *body == "" {
		return 0
	}
	return len(strings.Fields(*body))
}

// computeReadingTime returns estimated reading time in minutes (min 1 if content exists).
// computeReadingTime 返回预计阅读分钟数（有内容时最少 1）。
func computeReadingTime(wordCount int) int {
	if wordCount == 0 {
		return 0
	}
	t := wordCount / 200
	if t < 1 {
		t = 1
	}
	return t
}

// validStatusTransition checks if a status transition is allowed.
// validStatusTransition 检查状态流转是否允许。
func validStatusTransition(from, to string) bool {
	allowed := map[string][]string{
		"draft":     {"review", "archived"},
		"review":    {"published", "draft"},
		"published": {"archived"},
		"archived":  {"draft"},
	}
	targets, ok := allowed[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// mapContentCreateUniqueViolation maps a unique-constraint violation on content create.
// mapContentCreateUniqueViolation 映射内容创建时的唯一约束冲突。
func mapContentCreateUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return common.NewConflict("content slug is already taken for this type", err)
	}
	return common.NewInternal("failed to create content", err)
}

// mapContentUpdateUniqueViolation maps a unique-constraint violation on content update.
// mapContentUpdateUniqueViolation 映射内容更新时的唯一约束冲突。
func mapContentUpdateUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return common.NewConflict("content slug is already taken for this type", err)
	}
	return common.NewInternal("failed to update content", err)
}

// mapVersionUniqueViolation maps a unique-constraint violation on version create.
// mapVersionUniqueViolation 映射版本创建时的唯一约束冲突。
func mapVersionUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return common.NewConflict("version number conflict; retry the request", err)
	}
	return common.NewInternal("failed to create version", err)
}
