package archives

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"k8s.io/klog/v2"
)

// BatchArticles 批量操作文章。
// 支持 action：delete / set_status / set_category / set_tags。
func (a *ArticleAdmin) BatchArticles(ctx context.Context, req *v1.BatchArticleRequest) (*v1.BatchArticleResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	ids := deduplicateIDs(req.IDs)
	if len(ids) == 0 {
		return nil, http.StatusBadRequest, errors.New("ids must not be empty")
	}

	switch req.Action {
	case "delete":
		return a.batchDelete(ctx, ids)
	case "set_status":
		return a.batchSetStatus(ctx, ids, req.Payload.Status)
	case "set_category":
		return a.batchSetCategory(ctx, ids, req.Payload.CategoryID)
	case "set_tags":
		return a.batchSetTags(ctx, ids, req.Payload.TagIDs)
	default:
		return nil, http.StatusBadRequest, errors.New("unsupported action")
	}
}

func deduplicateIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (a *ArticleAdmin) batchDelete(ctx context.Context, ids []int64) (*v1.BatchArticleResponse, int, error) {
	now := time.Now()
	res := a.svc.DB.WithContext(ctx).
		Model(&models.Article{}).
		Where("id IN ? AND deleted_at IS NULL", ids).
		Update("deleted_at", now)
	if res.Error != nil {
		klog.ErrorS(res.Error, "batchDelete")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	// 清理关联标签
	_ = a.svc.DB.WithContext(ctx).Where("article_id IN ?", ids).Delete(&models.ArticleTag{})
	return &v1.BatchArticleResponse{Affected: res.RowsAffected}, http.StatusOK, nil
}

func (a *ArticleAdmin) batchSetStatus(ctx context.Context, ids []int64, status string) (*v1.BatchArticleResponse, int, error) {
	status = strings.TrimSpace(status)
	validStatuses := map[string]bool{
		"draft": true, "published": true, "archived": true, "private": true, "scheduled": true,
	}
	if !validStatuses[status] {
		return nil, http.StatusBadRequest, errors.New("invalid status value")
	}
	updates := map[string]interface{}{"status": status}
	if models.ArticleStatus(status) == models.ArticleStatusPublished {
		now := time.Now()
		updates["published_at"] = now
	}
	res := a.svc.DB.WithContext(ctx).
		Model(&models.Article{}).
		Where("id IN ? AND deleted_at IS NULL", ids).
		Updates(updates)
	if res.Error != nil {
		klog.ErrorS(res.Error, "batchSetStatus")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return &v1.BatchArticleResponse{Affected: res.RowsAffected}, http.StatusOK, nil
}

func (a *ArticleAdmin) batchSetCategory(ctx context.Context, ids []int64, categoryID *int64) (*v1.BatchArticleResponse, int, error) {
	catID := normalizeOptionalCategoryID(categoryID)
	if catID != nil {
		if err := a.validateCategory(ctx, catID); err != nil {
			if err.Error() == "category not found" {
				return nil, http.StatusBadRequest, err
			}
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}
	res := a.svc.DB.WithContext(ctx).
		Model(&models.Article{}).
		Where("id IN ? AND deleted_at IS NULL", ids).
		Update("category_id", catID)
	if res.Error != nil {
		klog.ErrorS(res.Error, "batchSetCategory")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return &v1.BatchArticleResponse{Affected: res.RowsAffected}, http.StatusOK, nil
}

func (a *ArticleAdmin) batchSetTags(ctx context.Context, ids []int64, tagIDs []int64) (*v1.BatchArticleResponse, int, error) {
	tagIDsNorm := normalizeTagIDs(tagIDs)
	if err := a.validateTags(ctx, tagIDsNorm); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, http.StatusBadRequest, err
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	tx := a.svc.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	for _, articleID := range ids {
		if err := a.replaceTags(ctx, tx, articleID, tagIDsNorm); err != nil {
			tx.Rollback()
			klog.ErrorS(err, "batchSetTags: replaceTags", "articleID", articleID)
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return &v1.BatchArticleResponse{Affected: int64(len(ids))}, http.StatusOK, nil
}
