package archives

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

const maxVersionsKept = 50

// saveVersion 在事务中为文章保存当前快照为新版本记录。
// 调用方须在事务提交前调用（tx 为事务 DB）。
func (a *ArticleAdmin) saveVersion(ctx context.Context, tx *gorm.DB, art *models.Article, operatorID int64) error {
	var maxVer int
	if err := tx.WithContext(ctx).Model(&models.ArticleVersion{}).
		Where("article_id = ?", art.ID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVer).Error; err != nil {
		return err
	}
	ver := &models.ArticleVersion{
		ArticleID: art.ID,
		Title:     art.Title,
		Content:   art.Content,
		Version:   maxVer + 1,
		CreatedBy: operatorID,
	}
	if err := tx.WithContext(ctx).Create(ver).Error; err != nil {
		return err
	}
	// 超过上限时删除最旧的版本（保留最新 maxVersionsKept 条）
	if maxVer+1 > maxVersionsKept {
		if err := tx.WithContext(ctx).
			Where("article_id = ? AND version <= ?", art.ID, maxVer+1-maxVersionsKept).
			Delete(&models.ArticleVersion{}).Error; err != nil {
			klog.Warningf("saveVersion: trim old versions: %v", err)
		}
	}
	return nil
}

// ListVersions 列出文章版本历史（最多 maxVersionsKept 条，按 version DESC）。
func (a *ArticleAdmin) ListVersions(ctx context.Context, articleID int64) (*v1.ArticleVersionListResponse, int, error) {
	if articleID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid article id")
	}
	var exists int64
	if err := a.svc.DB.WithContext(ctx).Model(&models.Article{}).
		Where("id = ? AND deleted_at IS NULL", articleID).Count(&exists).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if exists == 0 {
		return nil, http.StatusNotFound, errors.New("article not found")
	}

	var rows []models.ArticleVersion
	if err := a.svc.DB.WithContext(ctx).
		Where("article_id = ?", articleID).
		Order("version DESC").
		Limit(maxVersionsKept).
		Find(&rows).Error; err != nil {
		klog.ErrorS(err, "ListVersions: find", "articleID", articleID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	items := make([]v1.ArticleVersionItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, v1.ArticleVersionItem{
			ID:        r.ID,
			ArticleID: r.ArticleID,
			Title:     r.Title,
			Version:   r.Version,
			CreatedBy: r.CreatedBy,
			CreatedAt: r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return &v1.ArticleVersionListResponse{Items: items, Total: len(items)}, http.StatusOK, nil
}

// RestoreVersion 将指定版本的 title/content 写回文章，并将当前内容保存为新版本。
func (a *ArticleAdmin) RestoreVersion(ctx context.Context, articleID, versionID, operatorID int64) (*v1.ArticleDetailResponse, int, error) {
	if articleID <= 0 || versionID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid id")
	}

	var art models.Article
	if err := a.svc.DB.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", articleID).First(&art).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("article not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	var ver models.ArticleVersion
	if err := a.svc.DB.WithContext(ctx).
		Where("id = ? AND article_id = ?", versionID, articleID).First(&ver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("version not found")
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

	// 先保存当前内容为新版本
	if err := a.saveVersion(ctx, tx, &art, operatorID); err != nil {
		tx.Rollback()
		klog.ErrorS(err, "RestoreVersion: saveVersion", "articleID", articleID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	// 写回历史版本内容
	if err := tx.Model(&models.Article{}).Where("id = ?", articleID).
		Updates(map[string]interface{}{
			"title":   ver.Title,
			"content": ver.Content,
		}).Error; err != nil {
		tx.Rollback()
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	maybeHexoSyncSingle(a.svc, articleID)
	return a.loadArticleDetail(ctx, articleID)
}
