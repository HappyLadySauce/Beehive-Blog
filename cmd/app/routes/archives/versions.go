package archives

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	routehexo "github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/hexo"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

const maxVersionsKept = 50

// autosaveVersionNumber 自动保存槽固定版本号，不参与手动版递增。
const autosaveVersionNumber = 0

// saveVersion 在事务中为文章保存当前快照为新版本记录（仅手动版本，不含自动保存槽）。
// 调用方须在事务提交前调用（tx 为事务 DB）。
func (a *ArticleAdmin) saveVersion(ctx context.Context, tx *gorm.DB, art *models.Article, operatorID int64) error {
	var maxVer int
	if err := tx.WithContext(ctx).Model(&models.ArticleVersion{}).
		Where("article_id = ? AND is_autosave = ?", art.ID, false).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVer).Error; err != nil {
		return err
	}
	ver := &models.ArticleVersion{
		ArticleID:  art.ID,
		Title:      art.Title,
		Content:    art.Content,
		Version:    maxVer + 1,
		IsAutosave: false,
		CreatedBy:  operatorID,
	}
	if err := tx.WithContext(ctx).Create(ver).Error; err != nil {
		return err
	}
	newVer := maxVer + 1
	if newVer > maxVersionsKept {
		threshold := newVer - maxVersionsKept
		if err := tx.WithContext(ctx).
			Where("article_id = ? AND is_autosave = ? AND version <= ?", art.ID, false, threshold).
			Delete(&models.ArticleVersion{}).Error; err != nil {
			klog.Warningf("saveVersion: trim old versions: %v", err)
		}
	}
	return nil
}

// saveOrReplaceAutosaveSnapshot 将当前正文快照写入「自动保存」单槽：已存在则覆盖，否则插入。
func (a *ArticleAdmin) saveOrReplaceAutosaveSnapshot(ctx context.Context, tx *gorm.DB, art *models.Article, operatorID int64) error {
	var existing models.ArticleVersion
	err := tx.WithContext(ctx).Where("article_id = ? AND is_autosave = ?", art.ID, true).First(&existing).Error
	now := time.Now()
	if err == nil {
		return tx.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"title":      art.Title,
			"content":    art.Content,
			"created_by": operatorID,
			"created_at": now,
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	ver := &models.ArticleVersion{
		ArticleID:  art.ID,
		Title:      art.Title,
		Content:    art.Content,
		Version:    autosaveVersionNumber,
		IsAutosave: true,
		CreatedBy:  operatorID,
		CreatedAt:  now,
	}
	return tx.WithContext(ctx).Create(ver).Error
}

// ListVersions 列出文章版本历史（手动版本最多 maxVersionsKept 条 + 至多一条自动保存，手动在前、自动保存殿后）。
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
		Order("is_autosave ASC, version DESC").
		Limit(maxVersionsKept + 1).
		Find(&rows).Error; err != nil {
		klog.ErrorS(err, "ListVersions: find", "articleID", articleID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	items := make([]v1.ArticleVersionItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, v1.ArticleVersionItem{
			ID:         r.ID,
			ArticleID:  r.ArticleID,
			Title:      r.Title,
			Version:    r.Version,
			IsAutosave: r.IsAutosave,
			CreatedBy:  r.CreatedBy,
			CreatedAt:  r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return &v1.ArticleVersionListResponse{Items: items, Total: len(items)}, http.StatusOK, nil
}

// UpdateArticleVersion 更新指定版本记录的标题（展示名），不修改正文快照。
func (a *ArticleAdmin) UpdateArticleVersion(ctx context.Context, articleID, versionID int64, req *v1.UpdateArticleVersionRequest) (*v1.ArticleVersionItem, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	if articleID <= 0 || versionID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid id")
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, http.StatusBadRequest, errors.New("invalid title")
	}

	var artExists int64
	if err := a.svc.DB.WithContext(ctx).Model(&models.Article{}).
		Where("id = ? AND deleted_at IS NULL", articleID).Count(&artExists).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if artExists == 0 {
		return nil, http.StatusNotFound, errors.New("article not found")
	}

	res := a.svc.DB.WithContext(ctx).Model(&models.ArticleVersion{}).
		Where("id = ? AND article_id = ?", versionID, articleID).
		Update("title", title)
	if res.Error != nil {
		klog.ErrorS(res.Error, "UpdateArticleVersion", "articleID", articleID, "versionID", versionID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if res.RowsAffected == 0 {
		return nil, http.StatusNotFound, errors.New("version not found")
	}

	var row models.ArticleVersion
	if err := a.svc.DB.WithContext(ctx).
		Where("id = ? AND article_id = ?", versionID, articleID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("version not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	item := v1.ArticleVersionItem{
		ID:         row.ID,
		ArticleID:  row.ArticleID,
		Title:      row.Title,
		Version:    row.Version,
		IsAutosave: row.IsAutosave,
		CreatedBy:  row.CreatedBy,
		CreatedAt:  row.CreatedAt.UTC().Format(time.RFC3339),
	}
	return &item, http.StatusOK, nil
}

// DeleteArticleVersion 硬删除一条版本记录。
func (a *ArticleAdmin) DeleteArticleVersion(ctx context.Context, articleID, versionID int64) (*v1.DeleteArticleVersionResponse, int, error) {
	if articleID <= 0 || versionID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid id")
	}

	var artExists int64
	if err := a.svc.DB.WithContext(ctx).Model(&models.Article{}).
		Where("id = ? AND deleted_at IS NULL", articleID).Count(&artExists).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if artExists == 0 {
		return nil, http.StatusNotFound, errors.New("article not found")
	}

	res := a.svc.DB.WithContext(ctx).
		Where("id = ? AND article_id = ?", versionID, articleID).
		Delete(&models.ArticleVersion{})
	if res.Error != nil {
		klog.ErrorS(res.Error, "DeleteArticleVersion", "articleID", articleID, "versionID", versionID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if res.RowsAffected == 0 {
		return nil, http.StatusNotFound, errors.New("version not found")
	}
	return &v1.DeleteArticleVersionResponse{ID: versionID}, http.StatusOK, nil
}

// RestoreVersion 将指定历史版本的 title/content 写回文章（恢复前不额外生成新版本快照）。
func (a *ArticleAdmin) RestoreVersion(ctx context.Context, articleID, versionID, _ int64) (*v1.ArticleDetailResponse, int, error) {
	if articleID <= 0 || versionID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid id")
	}

	var artExists int64
	if err := a.svc.DB.WithContext(ctx).Model(&models.Article{}).
		Where("id = ? AND deleted_at IS NULL", articleID).Count(&artExists).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if artExists == 0 {
		return nil, http.StatusNotFound, errors.New("article not found")
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

	routehexo.MaybeSyncArticle(a.svc, articleID)
	var row models.Article
	if err := a.svc.DB.WithContext(ctx).Where("id = ?", articleID).First(&row).Error; err == nil {
		a.syncArticleAttachmentRefs(ctx, articleID, row.Content, row.Summary)
	}
	return a.loadArticleDetail(ctx, articleID)
}
