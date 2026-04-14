package attachments

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/v1/pkg/attachmentref"
	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/disintegration/imaging"
	gwebp "github.com/gen2brain/webp"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	_ "golang.org/x/image/webp" // register webp decoder
)

const settingKeyAttachmentProcessing = "attachment.processing"

func defaultProcessingSettings() v1.AttachmentProcessingSettings {
	return v1.AttachmentProcessingSettings{
		DefaultQuality: 85,
		AllowedFormats: []string{"jpeg", "png", "gif", "webp", "ico"},
	}
}

// GetProcessingSettings 读取全局附件处理默认配置。
func (s *Service) GetProcessingSettings(ctx context.Context) (*v1.AttachmentProcessingSettings, int, error) {
	var st models.Setting
	err := s.svc.DB.WithContext(ctx).Where("key = ?", settingKeyAttachmentProcessing).First(&st).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			d := defaultProcessingSettings()
			return &d, http.StatusOK, nil
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	var out v1.AttachmentProcessingSettings
	if st.Value != "" {
		if err := json.Unmarshal([]byte(st.Value), &out); err != nil {
			d := defaultProcessingSettings()
			return &d, http.StatusOK, nil
		}
		if len(out.AllowedFormats) == 0 {
			out.AllowedFormats = defaultProcessingSettings().AllowedFormats
		}
		return &out, http.StatusOK, nil
	}
	d := defaultProcessingSettings()
	return &d, http.StatusOK, nil
}

// PutProcessingSettings 保存全局附件处理默认配置。
func (s *Service) PutProcessingSettings(ctx context.Context, req *v1.AttachmentProcessingSettings) (*v1.AttachmentProcessingSettings, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	b, err := json.Marshal(req)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	var existing models.Setting
	e := s.svc.DB.WithContext(ctx).Where("key = ?", settingKeyAttachmentProcessing).First(&existing).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		st := models.Setting{Key: settingKeyAttachmentProcessing, Value: string(b), Group: models.SettingGroupGeneral}
		if err := s.svc.DB.WithContext(ctx).Create(&st).Error; err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	} else if e != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	} else {
		if err := s.svc.DB.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"value": string(b),
			"group": models.SettingGroupGeneral,
		}).Error; err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}
	return req, http.StatusOK, nil
}

// Family 返回根附件、派生子附件，以及各家族成员在 article_attachments 中的引用文章列表。
func (s *Service) Family(ctx context.Context, id int64) (*v1.AttachmentFamilyResponse, int, error) {
	if id <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid id")
	}
	var a models.Attachment
	if err := s.svc.DB.WithContext(ctx).Preload("Group").Where("id = ?", id).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("attachment not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	rootID := a.ID
	if a.ParentID != nil {
		rootID = *a.ParentID
	}
	var root models.Attachment
	if err := s.svc.DB.WithContext(ctx).Preload("Group").Where("id = ?", rootID).First(&root).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	var children []models.Attachment
	if err := s.svc.DB.WithContext(ctx).Preload("Group").Where("parent_id = ?", rootID).Order("created_at ASC").Find(&children).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	rootRef := s.refCount(ctx, root.ID)
	out := &v1.AttachmentFamilyResponse{
		Root:             s.toItem(&root, rootRef),
		Children:         make([]v1.AttachmentItem, 0, len(children)),
		MemberReferences: []v1.AttachmentMemberArticleRefs{},
	}
	for i := range children {
		out.Children = append(out.Children, s.toItem(&children[i], s.refCount(ctx, children[i].ID)))
	}

	familyIDs := []int64{root.ID}
	for i := range children {
		familyIDs = append(familyIDs, children[i].ID)
	}
	type aaPair struct {
		AttachmentID int64 `gorm:"column:attachment_id"`
		ArticleID    int64 `gorm:"column:article_id"`
	}
	var pairs []aaPair
	if len(familyIDs) > 0 {
		if err := s.svc.DB.WithContext(ctx).Model(&models.ArticleAttachment{}).
			Select("attachment_id", "article_id").
			Where("attachment_id IN ?", familyIDs).
			Scan(&pairs).Error; err != nil {
			klog.ErrorS(err, "Family: article_attachments pairs")
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}
	byAtt := make(map[int64]map[int64]struct{})
	for _, fid := range familyIDs {
		byAtt[fid] = make(map[int64]struct{})
	}
	for _, p := range pairs {
		if m, ok := byAtt[p.AttachmentID]; ok {
			m[p.ArticleID] = struct{}{}
		}
	}
	allArtIDs := make([]int64, 0)
	seenArt := make(map[int64]struct{})
	for _, m := range byAtt {
		for aid := range m {
			if _, ok := seenArt[aid]; ok {
				continue
			}
			seenArt[aid] = struct{}{}
			allArtIDs = append(allArtIDs, aid)
		}
	}
	artByID := make(map[int64]models.Article)
	if len(allArtIDs) > 0 {
		var arts []models.Article
		if err := s.svc.DB.WithContext(ctx).Where("id IN ?", allArtIDs).Find(&arts).Error; err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		for i := range arts {
			artByID[arts[i].ID] = arts[i]
		}
	}
	buildRefs := func(articleIDSet map[int64]struct{}) []v1.AttachmentFamilyArticleRef {
		if len(articleIDSet) == 0 {
			return nil
		}
		ids := make([]int64, 0, len(articleIDSet))
		for aid := range articleIDSet {
			ids = append(ids, aid)
		}
		var refs []v1.AttachmentFamilyArticleRef
		for _, aid := range ids {
			if ar, ok := artByID[aid]; ok {
				refs = append(refs, v1.AttachmentFamilyArticleRef{
					ArticleID: ar.ID,
					Title:     ar.Title,
					Slug:      ar.Slug,
				})
			}
		}
		sort.Slice(refs, func(i, j int) bool {
			return artByID[refs[i].ArticleID].UpdatedAt.After(artByID[refs[j].ArticleID].UpdatedAt)
		})
		return refs
	}
	out.MemberReferences = append(out.MemberReferences, v1.AttachmentMemberArticleRefs{
		AttachmentID: root.ID,
		Articles:     buildRefs(byAtt[root.ID]),
	})
	for i := range children {
		cid := children[i].ID
		out.MemberReferences = append(out.MemberReferences, v1.AttachmentMemberArticleRefs{
			AttachmentID: cid,
			Articles:     buildRefs(byAtt[cid]),
		})
	}
	return out, http.StatusOK, nil
}

func (s *Service) refCount(ctx context.Context, attachmentID int64) int64 {
	var n int64
	_ = s.svc.DB.WithContext(ctx).Model(&models.ArticleAttachment{}).
		Where("attachment_id = ?", attachmentID).Count(&n).Error
	return n
}

func (s *Service) toItem(a *models.Attachment, ref int64) v1.AttachmentItem {
	gn := ""
	if a.Group != nil {
		gn = a.Group.Name
	}
	return v1.AttachmentItem{
		ID:           a.ID,
		Name:         a.Name,
		OriginalName: a.OriginalName,
		URL:          a.URL,
		ThumbURL:     a.ThumbURL,
		Type:         string(a.Type),
		MimeType:     a.MimeType,
		Size:         a.Size,
		Width:        a.Width,
		Height:       a.Height,
		ParentID:     a.ParentID,
		Variant:      a.Variant,
		GroupID:      a.GroupID,
		GroupName:    gn,
		RefCount:     ref,
		CreatedAt:    a.CreatedAt,
	}
}

func normalizeImageFormat(f string) string {
	f = strings.ToLower(strings.TrimSpace(f))
	if f == "jpg" {
		return "jpeg"
	}
	return f
}

// detectOutputFormat 在 format 为空时根据根附件推断输出格式。
func detectOutputFormat(root *models.Attachment) string {
	mt := strings.ToLower(strings.TrimSpace(root.MimeType))
	switch {
	case strings.Contains(mt, "png"):
		return "png"
	case strings.Contains(mt, "gif"):
		return "gif"
	case strings.Contains(mt, "webp"):
		return "webp"
	case strings.Contains(mt, "jpeg"), strings.Contains(mt, "jpg"):
		return "jpeg"
	case strings.Contains(mt, "icon"), strings.Contains(mt, "ico"):
		return "ico"
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(root.Name), "."))
	switch ext {
	case "png", "gif", "webp", "ico":
		return ext
	case "jpg", "jpeg":
		return "jpeg"
	}
	return "jpeg"
}

// attachmentStemForDerivatives 用于派生文件名的词干：优先用户上传时的原始文件名，与列表展示一致。
func attachmentStemForDerivatives(root *models.Attachment) string {
	base := strings.TrimSpace(root.OriginalName)
	if base == "" {
		base = root.Name
	}
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	return safeFilename(stem)
}

func extForFormat(format string) string {
	switch format {
	case "jpeg":
		return ".jpg"
	case "png":
		return ".png"
	case "gif":
		return ".gif"
	case "webp":
		return ".webp"
	case "ico":
		return ".ico"
	default:
		return ".jpg"
	}
}

func mimeForFormat(format string) string {
	switch format {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "ico":
		return "image/x-icon"
	default:
		return "image/jpeg"
	}
}

func encodeRaster(format string, quality int, src image.Image) ([]byte, error) {
	var buf bytes.Buffer
	var err error
	switch format {
	case "jpeg":
		if quality <= 0 {
			quality = 85
		}
		err = jpeg.Encode(&buf, src, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(&buf, src)
	case "gif":
		err = gif.Encode(&buf, src, nil)
	case "webp":
		if quality <= 0 {
			quality = 85
		}
		err = gwebp.Encode(&buf, src, gwebp.Options{Quality: quality, Lossless: false})
	case "ico":
		icon := imaging.Resize(src, 256, 256, imaging.Lanczos)
		err = ico.Encode(&buf, icon)
	default:
		return nil, errors.New("unsupported format")
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// resolveAttachmentRootID 返回附件所在家族的根附件 ID。
func (s *Service) resolveAttachmentRootID(ctx context.Context, id int64) (int64, error) {
	var a models.Attachment
	if err := s.svc.DB.WithContext(ctx).Select("id", "parent_id").Where("id = ?", id).First(&a).Error; err != nil {
		return 0, err
	}
	if a.ParentID == nil {
		return a.ID, nil
	}
	return *a.ParentID, nil
}

func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// collectArticleIDsForURLReplace 关联表中的文章，以及正文/摘要中出现旧 URL 的文章（去重）。
func (s *Service) collectArticleIDsForURLReplace(ctx context.Context, tx *gorm.DB, attachmentID int64, oldURL string) ([]int64, error) {
	var junction []int64
	if err := tx.Model(&models.ArticleAttachment{}).Where("attachment_id = ?", attachmentID).Distinct().Pluck("article_id", &junction).Error; err != nil {
		return nil, err
	}
	seen := make(map[int64]struct{}, len(junction)+8)
	for _, id := range junction {
		if id > 0 {
			seen[id] = struct{}{}
		}
	}
	if strings.TrimSpace(oldURL) != "" {
		pat := "%" + escapeLikePattern(oldURL) + "%"
		var extra []int64
		if err := tx.Model(&models.Article{}).Where("content LIKE ? OR summary LIKE ?", pat, pat).Pluck("id", &extra).Error; err != nil {
			return nil, err
		}
		for _, id := range extra {
			if id > 0 {
				seen[id] = struct{}{}
			}
		}
	}
	out := make([]int64, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out, nil
}

// replaceOldURLInArticles 将正文/摘要中的旧 URL 候选替换为新 URL，并同步 attachment 引用索引。
func (s *Service) replaceOldURLInArticles(ctx context.Context, tx *gorm.DB, articleIDs []int64, oldURL, newURL, baseURL string) error {
	oldURLs := uniqueURLCandidates(oldURL)
	for _, aid := range articleIDs {
		if aid <= 0 {
			continue
		}
		var art models.Article
		if err := tx.Where("id = ?", aid).First(&art).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}
		content := art.Content
		summary := art.Summary
		for _, old := range oldURLs {
			if old == "" {
				continue
			}
			content = strings.ReplaceAll(content, old, newURL)
			summary = strings.ReplaceAll(summary, old, newURL)
		}
		if content == art.Content && summary == art.Summary {
			continue
		}
		if err := tx.Model(&models.Article{}).Where("id = ?", aid).Updates(map[string]interface{}{
			"content": content,
			"summary": summary,
		}).Error; err != nil {
			return err
		}
		if err := attachmentref.SyncForArticle(ctx, tx, aid, content, summary, baseURL); err != nil {
			return err
		}
	}
	return nil
}

// UpdateAttachment 可选物理重命名（同步 path/url 与正文引用）与/或更新分类；根附件变更分类时同步子附件 group_id。
func (s *Service) UpdateAttachment(ctx context.Context, id int64, req *v1.UpdateAttachmentRequest) (*v1.AttachmentItem, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	if req.Name == nil && req.GroupID == nil {
		return nil, http.StatusBadRequest, errors.New("name or groupId required")
	}

	var a models.Attachment
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", id).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("attachment not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	uploadDir := filepath.Clean(s.svc.Config.StorageOptions.UploadDir)
	baseURL := strings.TrimRight(strings.TrimSpace(s.svc.Config.StorageOptions.BaseURL), "/")
	if baseURL == "" {
		return nil, http.StatusInternalServerError, errors.New("storage not configured")
	}
	if a.Path == "" {
		return nil, http.StatusBadRequest, errors.New("missing file path")
	}

	oldPath := filepath.Clean(a.Path)
	relOld, err := filepath.Rel(uploadDir, oldPath)
	if err != nil || strings.HasPrefix(relOld, "..") {
		return nil, http.StatusBadRequest, errors.New("invalid attachment path")
	}

	oldURL := a.URL
	newPath := oldPath
	newURL := oldURL
	newName := a.Name
	var needDiskMove bool

	if req.Name != nil {
		raw := strings.TrimSpace(*req.Name)
		if raw == "" {
			return nil, http.StatusBadRequest, errors.New("invalid name")
		}
		curExt := filepath.Ext(a.Name)
		clean := safeFilename(raw)
		if clean == "" {
			return nil, http.StatusBadRequest, errors.New("invalid name")
		}
		var stem string
		if ext := filepath.Ext(clean); ext != "" {
			if !strings.EqualFold(ext, curExt) {
				return nil, http.StatusBadRequest, errors.New("file extension must match current attachment")
			}
			stem = strings.TrimSuffix(clean, ext)
		} else {
			stem = clean
		}
		if stem == "" {
			return nil, http.StatusBadRequest, errors.New("invalid name")
		}
		newName = stem + curExt
		dir := filepath.Dir(oldPath)
		newPath = filepath.Clean(filepath.Join(dir, newName))
		relNew, errRel := filepath.Rel(uploadDir, newPath)
		if errRel != nil || strings.HasPrefix(relNew, "..") {
			return nil, http.StatusBadRequest, errors.New("invalid target path")
		}
		newURL = baseURL + "/" + filepath.ToSlash(relNew)
		needDiskMove = oldPath != newPath
	}

	var rollback func()
	if needDiskMove {
		if st, err := os.Stat(newPath); err == nil && !st.IsDir() {
			return nil, http.StatusConflict, errors.New("target file already exists")
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			klog.ErrorS(err, "UpdateAttachment: rename", "from", oldPath, "to", newPath)
			return nil, http.StatusInternalServerError, errors.New("rename failed")
		}
		rollback = func() {
			if err := os.Rename(newPath, oldPath); err != nil {
				klog.ErrorS(err, "UpdateAttachment: rollback rename", "from", newPath, "to", oldPath)
			}
		}
	}

	err = s.svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.Name != nil {
			thumb := a.ThumbURL
			if thumb != "" && strings.Contains(thumb, oldURL) {
				thumb = strings.ReplaceAll(thumb, oldURL, newURL)
			}
			updates := map[string]interface{}{
				"name":          newName,
				"original_name": newName,
				"path":          newPath,
				"url":           newURL,
			}
			if a.ThumbURL != "" {
				updates["thumb_url"] = thumb
			}
			if err := tx.Model(&models.Attachment{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
			if oldURL != newURL {
				ids, err := s.collectArticleIDsForURLReplace(ctx, tx, id, oldURL)
				if err != nil {
					return err
				}
				if err := s.replaceOldURLInArticles(ctx, tx, ids, oldURL, newURL, baseURL); err != nil {
					return err
				}
			}
		}

		if req.GroupID != nil {
			var newGID *int64
			if *req.GroupID <= 0 {
				newGID = nil
			} else {
				gid := *req.GroupID
				var n int64
				if err := tx.Model(&models.AttachmentGroup{}).Where("id = ?", gid).Count(&n).Error; err != nil {
					return err
				}
				if n == 0 {
					return errAttachmentGroupNotFound
				}
				newGID = &gid
			}
			if a.ParentID == nil {
				if err := tx.Model(&models.Attachment{}).Where("id = ? OR parent_id = ?", a.ID, a.ID).Update("group_id", newGID).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Model(&models.Attachment{}).Where("id = ?", a.ID).Update("group_id", newGID).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		if rollback != nil {
			rollback()
		}
		if errors.Is(err, errAttachmentGroupNotFound) {
			return nil, http.StatusNotFound, errors.New("attachment group not found")
		}
		klog.ErrorS(err, "UpdateAttachment", "id", id)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	if err := s.svc.DB.WithContext(ctx).Preload("Group").Where("id = ?", id).First(&a).Error; err != nil {
		if rollback != nil {
			rollback()
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	it := s.toItem(&a, s.refCount(ctx, a.ID))
	return &it, http.StatusOK, nil
}

var errAttachmentGroupNotFound = errors.New("attachment group not found")

// ReplaceAttachmentInArticles 在选中正文中将 from 附件 URL 替换为 to（须同一根）。
func (s *Service) ReplaceAttachmentInArticles(ctx context.Context, req *v1.ReplaceAttachmentInArticlesRequest) (*v1.ReplaceAttachmentInArticlesResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	if req.FromAttachmentID == req.ToAttachmentID {
		return nil, http.StatusBadRequest, errors.New("from and to must differ")
	}
	var fromA, toA models.Attachment
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", req.FromAttachmentID).First(&fromA).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("attachment not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", req.ToAttachmentID).First(&toA).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("attachment not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	rootFrom, err := s.resolveAttachmentRootID(ctx, fromA.ID)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	rootTo, err := s.resolveAttachmentRootID(ctx, toA.ID)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if rootFrom != rootTo {
		return nil, http.StatusBadRequest, errors.New("attachments must belong to the same family")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(s.svc.Config.StorageOptions.BaseURL), "/")
	if baseURL == "" {
		return nil, http.StatusInternalServerError, errors.New("storage not configured")
	}

	oldURLs := uniqueURLCandidates(fromA.URL)
	newURL := toA.URL

	n := 0
	err = s.svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, aid := range req.ArticleIDs {
			if aid <= 0 {
				continue
			}
			var art models.Article
			if err := tx.Where("id = ?", aid).First(&art).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}
			content := art.Content
			summary := art.Summary
			for _, old := range oldURLs {
				if old == "" {
					continue
				}
				content = strings.ReplaceAll(content, old, newURL)
				summary = strings.ReplaceAll(summary, old, newURL)
			}
			if content == art.Content && summary == art.Summary {
				continue
			}
			if err := tx.Model(&models.Article{}).Where("id = ?", aid).Updates(map[string]interface{}{
				"content": content,
				"summary": summary,
			}).Error; err != nil {
				return err
			}
			if err := attachmentref.SyncForArticle(ctx, tx, aid, content, summary, baseURL); err != nil {
				return err
			}
			n++
		}
		return nil
	})
	if err != nil {
		klog.ErrorS(err, "ReplaceAttachmentInArticles")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return &v1.ReplaceAttachmentInArticlesResponse{Updated: n}, http.StatusOK, nil
}

func uniqueURLCandidates(raw string) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	add(raw)
	if n, err := attachmentref.NormalizeURLString(raw); err == nil {
		add(n)
	}
	return out
}

// ProcessImage 从根附件生成子附件（不修改原文件）；同名同参则覆盖已有派生文件与记录。
func (s *Service) ProcessImage(ctx context.Context, userID int64, rootOrChildID int64, req *v1.ProcessAttachmentRequest) (*v1.AttachmentItem, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	var a models.Attachment
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", rootOrChildID).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("attachment not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if a.Type != models.AttachmentTypeImage {
		return nil, http.StatusBadRequest, errors.New("only image attachments can be processed")
	}
	root := a
	if a.ParentID != nil {
		if err := s.svc.DB.WithContext(ctx).Where("id = ?", *a.ParentID).First(&root).Error; err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}
	if root.Path == "" {
		return nil, http.StatusBadRequest, errors.New("missing file path")
	}

	format := normalizeImageFormat(req.Format)
	if format == "" {
		format = detectOutputFormat(&root)
	}
	switch format {
	case "jpeg", "png", "gif", "webp", "ico":
	default:
		return nil, http.StatusBadRequest, errors.New("invalid format")
	}

	quality := req.Quality
	if quality <= 0 {
		quality = 85
	}

	src, err := imaging.Open(root.Path)
	if err != nil {
		klog.ErrorS(err, "ProcessImage: open", "path", root.Path)
		return nil, http.StatusBadRequest, errors.New("cannot decode source image")
	}

	data, err := encodeRaster(format, quality, src)
	if err != nil {
		klog.ErrorS(err, "ProcessImage: encode", "format", format)
		return nil, http.StatusInternalServerError, errors.New("encode failed")
	}

	policy, err := s.defaultPolicy(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("storage not configured")
	}
	uploadDir := filepath.Clean(s.svc.Config.StorageOptions.UploadDir)
	baseURL := strings.TrimRight(s.svc.Config.StorageOptions.BaseURL, "/")

	stem := attachmentStemForDerivatives(&root)
	ext := extForFormat(format)
	outFileName := fmt.Sprintf("%s-compress%d%s", stem, quality, ext)

	var existing models.Attachment
	q := s.svc.DB.WithContext(ctx).Where("parent_id = ? AND name = ?", root.ID, outFileName)
	err = q.First(&existing).Error
	isUpdate := err == nil
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	var diskPath string
	var relPath string
	now := time.Now()
	if isUpdate {
		diskPath = existing.Path
		rel, err := filepath.Rel(uploadDir, filepath.Clean(diskPath))
		if err != nil || strings.HasPrefix(rel, "..") {
			klog.ErrorS(err, "ProcessImage: rel path", "path", diskPath)
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		relPath = filepath.ToSlash(rel)
	} else {
		subDir := filepath.Join(uploadDir, fmt.Sprintf("%d/%02d", now.Year(), now.Month()))
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		diskPath = filepath.Join(subDir, outFileName)
		relPath = filepath.ToSlash(filepath.Join(fmt.Sprintf("%d/%02d", now.Year(), now.Month()), outFileName))
	}

	if err := os.WriteFile(diskPath, data, 0o644); err != nil {
		klog.ErrorS(err, "ProcessImage: write", "path", diskPath)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	w, h := 0, 0
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
		w, h = cfg.Width, cfg.Height
	}

	mimeType := mimeForFormat(format)
	fileURL := baseURL + "/" + relPath
	pid := root.ID
	srcExt := strings.TrimPrefix(strings.ToLower(filepath.Ext(root.Name)), ".")
	outExt := strings.TrimPrefix(ext, ".")
	if srcExt == "jpeg" {
		srcExt = "jpg"
	}
	if outExt == "jpeg" {
		outExt = "jpg"
	}
	variant := "compressed"
	if srcExt != outExt {
		variant = "converted"
	}

	var child models.Attachment
	if isUpdate {
		if err := s.svc.DB.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"path":      diskPath,
			"url":       fileURL,
			"mime_type": mimeType,
			"size":      int64(len(data)),
			"width":     w,
			"height":    h,
			"variant":   variant,
		}).Error; err != nil {
			_ = os.Remove(diskPath)
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if err := s.svc.DB.WithContext(ctx).Preload("Group").Where("id = ?", existing.ID).First(&child).Error; err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	} else {
		child = models.Attachment{
			Name:         outFileName,
			OriginalName: outFileName,
			Path:         diskPath,
			URL:          fileURL,
			Type:         models.AttachmentTypeImage,
			MimeType:     mimeType,
			Size:         int64(len(data)),
			Width:        w,
			Height:       h,
			PolicyID:     policy.ID,
			GroupID:      root.GroupID,
			ParentID:     &pid,
			Variant:      variant,
			UploadedBy:   userID,
		}
		if err := s.svc.DB.WithContext(ctx).Create(&child).Error; err != nil {
			_ = os.Remove(diskPath)
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if err := s.svc.DB.WithContext(ctx).Preload("Group").Where("id = ?", child.ID).First(&child).Error; err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}

	item := s.toItem(&child, s.refCount(ctx, child.ID))
	return &item, http.StatusOK, nil
}
