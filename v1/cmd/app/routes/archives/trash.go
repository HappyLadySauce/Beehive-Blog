package archives

import (
	"context"
	"errors"
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	routehexo "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/routes/hexo"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/v1/pkg/articlequery"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// ListTrashedArticles 分页列出回收站中的软删文章。
func (a *ArticleAdmin) ListTrashedArticles(ctx context.Context, req *v1.AdminArticleListRequest) (*v1.AdminArticleListResponse, int, error) {
	if req == nil {
		req = &v1.AdminArticleListRequest{}
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	db := a.svc.DB.WithContext(ctx)
	q := articlequery.TrashedArticleQuery(db, req.Keyword)
	rows, total, err := articlequery.ListAdminPage(ctx, db, q, page, pageSize, req.Sort)
	if err != nil {
		klog.ErrorS(err, "ListTrashedArticles query failed")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	list := make([]v1.AdminArticleListItem, 0, len(rows))
	for i := range rows {
		item := articlequery.MapListItem(&rows[i])
		list = append(list, v1.AdminArticleListItem{
			ArticleListItem: item,
			Status:          string(rows[i].Status),
		})
	}
	return &v1.AdminArticleListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK, nil
}

// RestoreArticle 从回收站恢复文章（清除 deleted_at）。
func (a *ArticleAdmin) RestoreArticle(ctx context.Context, articleID int64) (*v1.DeleteArticleResponse, int, error) {
	res := a.svc.DB.WithContext(ctx).Unscoped().Model(&models.Article{}).
		Where("id = ? AND deleted_at IS NOT NULL", articleID).
		Update("deleted_at", nil)
	if res.Error != nil {
		klog.ErrorS(res.Error, "RestoreArticle", "articleID", articleID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if res.RowsAffected == 0 {
		return nil, http.StatusNotFound, errors.New("article not found in trash")
	}
	routehexo.MaybeSyncArticle(a.svc, articleID)
	return &v1.DeleteArticleResponse{ID: articleID}, http.StatusOK, nil
}

// PermanentDeleteArticle 永久删除回收站中的文章（硬删）及关联 article_tags。
func (a *ArticleAdmin) PermanentDeleteArticle(ctx context.Context, articleID int64) (*v1.DeleteArticleResponse, int, error) {
	var n int64
	if err := a.svc.DB.WithContext(ctx).Unscoped().Model(&models.Article{}).
		Where("id = ? AND deleted_at IS NOT NULL", articleID).Count(&n).Error; err != nil {
		klog.ErrorS(err, "PermanentDeleteArticle count", "articleID", articleID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if n == 0 {
		return nil, http.StatusNotFound, errors.New("article not found in trash")
	}

	err := a.svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("article_id = ?", articleID).Delete(&models.ArticleTag{}).Error; err != nil {
			return err
		}
		res := tx.Unscoped().Where("id = ?", articleID).Delete(&models.Article{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("article not found in trash")
		}
		klog.ErrorS(err, "PermanentDeleteArticle", "articleID", articleID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	routehexo.MaybeDeletePost(a.svc, articleID)
	return &v1.DeleteArticleResponse{ID: articleID}, http.StatusOK, nil
}
