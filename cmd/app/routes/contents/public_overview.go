package contents

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// siteOverview returns the public reader shell data in a bounded query set.
// siteOverview 以有界查询集返回读者侧页面外壳数据。
func (c *ContentsController) siteOverview(ctx context.Context) (*v1.PublicSiteOverviewResponse, error) {
	latest, err := c.publicContentItems(ctx, publicContentQuery{Limit: 10})
	if err != nil {
		return nil, err
	}
	featured, err := c.publicContentItems(ctx, publicContentQuery{ContentType: "article", Limit: 4})
	if err != nil {
		return nil, err
	}
	categories, err := c.publicCategories(ctx, 12)
	if err != nil {
		return nil, err
	}
	tags, err := c.publicTags(ctx, 24)
	if err != nil {
		return nil, err
	}
	archives, err := c.publicArchives(ctx, 6)
	if err != nil {
		return nil, err
	}
	stats, err := c.publicStats(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.PublicSiteOverviewResponse{
		Latest:     latest,
		Featured:   featured,
		Recent:     latest,
		Categories: categories,
		Tags:       tags,
		Archives:   archives,
		Stats:      stats,
		Author: v1.SiteAuthor{
			Name:        "安和鱼",
			Description: "生活明朗，万物可爱",
		},
		GeneratedAt: time.Now().UTC(),
	}, nil
}

// SiteOverview handles GET /api/v1/public/site-overview.
// SiteOverview 处理 GET /api/v1/public/site-overview。
func (c *ContentsController) SiteOverview(ctx *gin.Context) {
	resp, err := c.siteOverview(ctx.Request.Context())
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

type publicContentQuery struct {
	ContentType string
	ExcludeID   int64
	Before      *time.Time
	After       *time.Time
	Limit       int
}

func (c *ContentsController) publicContentItems(ctx context.Context, in publicContentQuery) ([]v1.PublicContentItem, error) {
	limit := in.Limit
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	query := c.svc.DB.WithContext(ctx).Model(&model.Content{}).
		Where("status = ? AND visibility = ?", "published", "public")
	if in.ContentType != "" {
		query = query.Where("type = ?", in.ContentType)
	}
	if in.ExcludeID > 0 {
		query = query.Where("id <> ?", in.ExcludeID)
	}
	if in.Before != nil {
		query = query.Where("COALESCE(published_at, created_at) < ?", *in.Before)
	}
	if in.After != nil {
		query = query.Where("COALESCE(published_at, created_at) > ?", *in.After)
	}

	var rows []model.Content
	if err := query.Order("COALESCE(published_at, created_at) DESC, id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, common.NewInternal("failed to load public contents", fmt.Errorf("find: %w", err))
	}
	return c.decoratePublicContentItems(ctx, rows)
}

func (c *ContentsController) decoratePublicContentItems(ctx context.Context, rows []model.Content) ([]v1.PublicContentItem, error) {
	authorIDs := make([]int64, len(rows))
	contentIDs := make([]int64, len(rows))
	for i, row := range rows {
		authorIDs[i] = row.AuthorID
		contentIDs[i] = row.ID
	}
	authorMap, err := batchLoadAuthorUsernames(ctx, c, authorIDs)
	if err != nil {
		return nil, common.NewInternal("failed to load author usernames", err)
	}
	tagMap, err := batchLoadContentTags(ctx, c, contentIDs)
	if err != nil {
		return nil, common.NewInternal("failed to load content tags", err)
	}
	catMap, err := batchLoadContentCategories(ctx, c, contentIDs)
	if err != nil {
		return nil, common.NewInternal("failed to load content categories", err)
	}
	items := make([]v1.PublicContentItem, len(rows))
	for i, row := range rows {
		item := toPublicContentItem(row)
		item.AuthorUsername = authorMap[row.AuthorID]
		item.Tags = tagMap[row.ID]
		item.Categories = catMap[row.ID]
		attachPrimaryCategory(&item)
		items[i] = item
	}
	return items, nil
}

func (c *ContentsController) publicCategories(ctx context.Context, limit int) ([]v1.CategoryItem, error) {
	var cats []model.Category
	if err := c.svc.DB.WithContext(ctx).Order("sort_order ASC, id DESC").Limit(limit).Find(&cats).Error; err != nil {
		return nil, common.NewInternal("failed to load categories", err)
	}
	ids := make([]int64, len(cats))
	for i, cat := range cats {
		ids[i] = cat.ID
	}
	counts, err := publicCategoryContentCounts(ctx, c, ids)
	if err != nil {
		return nil, common.NewInternal("failed to load category counts", err)
	}
	items := categoriesToItems(cats)
	for i := range items {
		items[i].ContentCount = counts[items[i].ID]
	}
	return items, nil
}

func (c *ContentsController) publicTags(ctx context.Context, limit int) ([]v1.TagItem, error) {
	var tags []model.Tag
	if err := c.svc.DB.WithContext(ctx).Order("id DESC").Limit(limit).Find(&tags).Error; err != nil {
		return nil, common.NewInternal("failed to load tags", err)
	}
	ids := make([]int64, len(tags))
	for i, tag := range tags {
		ids[i] = tag.ID
	}
	counts, err := publicTagContentCounts(ctx, c, ids)
	if err != nil {
		return nil, common.NewInternal("failed to load tag counts", err)
	}
	items := tagsToItems(tags)
	for i := range items {
		items[i].ContentCount = counts[items[i].ID]
	}
	return items, nil
}

func (c *ContentsController) publicArchives(ctx context.Context, limit int) ([]v1.ArchiveItem, error) {
	type row struct {
		Year  int
		Month int
		Count int64
	}
	var rows []row
	if err := c.svc.DB.WithContext(ctx).Model(&model.Content{}).
		Select("EXTRACT(YEAR FROM COALESCE(published_at, created_at))::int AS year, EXTRACT(MONTH FROM COALESCE(published_at, created_at))::int AS month, COUNT(*) AS count").
		Where("status = ? AND visibility = ?", "published", "public").
		Group("year, month").
		Order("year DESC, month DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, common.NewInternal("failed to load archives", err)
	}
	items := make([]v1.ArchiveItem, len(rows))
	for i, row := range rows {
		items[i] = v1.ArchiveItem{Year: row.Year, Month: row.Month, Label: fmt.Sprintf("%d-%02d", row.Year, row.Month), Count: row.Count}
	}
	return items, nil
}

func (c *ContentsController) publicStats(ctx context.Context) (v1.SiteStats, error) {
	type row struct {
		Type  string
		Count int64
		Views int64
	}
	var rows []row
	if err := c.svc.DB.WithContext(ctx).Model(&model.Content{}).
		Select("type, COUNT(*) AS count, COALESCE(SUM(view_count), 0) AS views").
		Where("status = ? AND visibility = ?", "published", "public").
		Group("type").
		Find(&rows).Error; err != nil {
		return v1.SiteStats{}, common.NewInternal("failed to load site stats", err)
	}
	var stats v1.SiteStats
	for _, row := range rows {
		stats.Views += row.Views
		switch row.Type {
		case "article":
			stats.Articles = row.Count
		case "note":
			stats.Notes = row.Count
		case "project":
			stats.Projects = row.Count
		}
	}
	if err := c.svc.DB.WithContext(ctx).Model(&model.Tag{}).Count(&stats.Tags).Error; err != nil {
		return v1.SiteStats{}, common.NewInternal("failed to load tag count", err)
	}
	return stats, nil
}

func publicTagContentCounts(ctx context.Context, ctrl *ContentsController, tagIDs []int64) (map[int64]int64, error) {
	if len(tagIDs) == 0 {
		return nil, nil
	}
	type countRow struct {
		TagID int64
		Count int64
	}
	var rows []countRow
	if err := ctrl.svc.DB.WithContext(ctx).Table("content.content_tags AS ct").
		Select("ct.tag_id, COUNT(*) AS count").
		Joins("JOIN content.contents c ON c.id = ct.content_id AND c.status = ? AND c.visibility = ? AND c.deleted_at IS NULL", "published", "public").
		Where("ct.tag_id IN ?", tagIDs).
		Group("ct.tag_id").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[int64]int64, len(rows))
	for _, row := range rows {
		result[row.TagID] = row.Count
	}
	return result, nil
}

func publicCategoryContentCounts(ctx context.Context, ctrl *ContentsController, categoryIDs []int64) (map[int64]int64, error) {
	if len(categoryIDs) == 0 {
		return nil, nil
	}
	type countRow struct {
		CategoryID int64
		Count      int64
	}
	var rows []countRow
	if err := ctrl.svc.DB.WithContext(ctx).Table("content.content_categories AS cc").
		Select("cc.category_id, COUNT(*) AS count").
		Joins("JOIN content.contents c ON c.id = cc.content_id AND c.status = ? AND c.visibility = ? AND c.deleted_at IS NULL", "published", "public").
		Where("cc.category_id IN ?", categoryIDs).
		Group("cc.category_id").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[int64]int64, len(rows))
	for _, row := range rows {
		result[row.CategoryID] = row.Count
	}
	return result, nil
}
