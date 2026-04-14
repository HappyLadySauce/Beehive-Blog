package attachments

import (
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// Service 附件业务。
type Service struct {
	svc *svc.ServiceContext
}

// NewService constructs Service.
func NewService(svcCtx *svc.ServiceContext) *Service {
	return &Service{svc: svcCtx}
}

// defaultPolicy returns the default local StoragePolicy; cached on first call.
func (s *Service) defaultPolicy(ctx context.Context) (*models.StoragePolicy, error) {
	var p models.StoragePolicy
	if err := s.svc.DB.WithContext(ctx).
		Where("type = ? AND is_default = true", models.StoragePolicyLocal).
		First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("default storage policy not found; please run db migrations or check storage configuration")
		}
		return nil, err
	}
	return &p, nil
}

// detectAttachmentType returns the AttachmentType for a given MIME type string.
func detectAttachmentType(mimeType string) models.AttachmentType {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return models.AttachmentTypeImage
	case strings.HasPrefix(mimeType, "video/"):
		return models.AttachmentTypeVideo
	case strings.HasPrefix(mimeType, "audio/"):
		return models.AttachmentTypeAudio
	case mimeType == "application/pdf" ||
		strings.Contains(mimeType, "officedocument") ||
		strings.Contains(mimeType, "opendocument") ||
		mimeType == "application/msword" ||
		mimeType == "application/vnd.ms-excel" ||
		mimeType == "application/vnd.ms-powerpoint":
		return models.AttachmentTypeDocument
	default:
		return models.AttachmentTypeOther
	}
}

// sniffMimeType returns the MIME type for a multipart.FileHeader, preferring the
// Content-Type header and falling back to extension-based detection.
func sniffMimeType(fh *multipart.FileHeader) string {
	ct := fh.Header.Get("Content-Type")
	if ct != "" && ct != "application/octet-stream" {
		mediaType, _, err := mime.ParseMediaType(ct)
		if err == nil {
			return mediaType
		}
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if m := mime.TypeByExtension(ext); m != "" {
		mediaType, _, err := mime.ParseMediaType(m)
		if err == nil {
			return mediaType
		}
	}
	return "application/octet-stream"
}

// imageDimensions reads the image config (dimensions only) from an open file.
func imageDimensions(f multipart.File) (int, int) {
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

// safeFilename sanitises the original filename for storage (no directory traversal).
func safeFilename(original string) string {
	base := filepath.Base(original)
	// replace characters that are potentially problematic on various OSes
	replacer := strings.NewReplacer(" ", "_", "\t", "_", "\"", "", "'", "", "\\", "", "/", "")
	return replacer.Replace(base)
}

// Upload saves a multipart file to disk and writes a record to the database.
// onlyImage enforces that the MIME type must be image/*.
func (s *Service) Upload(ctx context.Context, userID int64, fh *multipart.FileHeader, groupID *int64, onlyImage bool) (*v1.AttachmentItem, int, error) {
	if fh == nil {
		return nil, http.StatusBadRequest, errors.New("no file provided")
	}
	maxSize := s.svc.Config.StorageOptions.MaxFileSize
	if fh.Size > maxSize {
		return nil, http.StatusBadRequest, fmt.Errorf("file too large (max %d bytes)", maxSize)
	}

	mimeType := sniffMimeType(fh)
	if onlyImage && !strings.HasPrefix(mimeType, "image/") {
		return nil, http.StatusBadRequest, errors.New("only image files are accepted by this endpoint")
	}
	attType := detectAttachmentType(mimeType)

	policy, err := s.defaultPolicy(ctx)
	if err != nil {
		klog.ErrorS(err, "Upload: get default policy")
		return nil, http.StatusInternalServerError, errors.New("storage not configured")
	}

	uploadDir := s.svc.Config.StorageOptions.UploadDir
	baseURL := strings.TrimRight(s.svc.Config.StorageOptions.BaseURL, "/")

	now := time.Now()
	subDir := filepath.Join(uploadDir, fmt.Sprintf("%d/%02d", now.Year(), now.Month()))
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		klog.ErrorS(err, "Upload: mkdirall", "dir", subDir)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	// generate a unique filename: timestamp_nanoseconds + extension
	uniqueBase := fmt.Sprintf("%d%d%s", now.Unix(), now.Nanosecond(), ext)
	diskPath := filepath.Join(subDir, uniqueBase)
	relPath := filepath.ToSlash(filepath.Join(fmt.Sprintf("%d/%02d", now.Year(), now.Month()), uniqueBase))

	src, err := fh.Open()
	if err != nil {
		klog.ErrorS(err, "Upload: open multipart file")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	defer src.Close()

	dst, err := os.Create(diskPath)
	if err != nil {
		klog.ErrorS(err, "Upload: create file", "path", diskPath)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	written, err := io.Copy(dst, src)
	dst.Close()
	if err != nil {
		_ = os.Remove(diskPath)
		klog.ErrorS(err, "Upload: write file", "path", diskPath)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	var width, height int
	if attType == models.AttachmentTypeImage {
		if rf, err := os.Open(diskPath); err == nil {
			width, height = imageDimensions(rf)
			rf.Close()
		}
	}

	fileURL := baseURL + "/" + relPath
	origName := safeFilename(fh.Filename)

	a := models.Attachment{
		Name:         uniqueBase,
		OriginalName: origName,
		Path:         diskPath,
		URL:          fileURL,
		Type:         attType,
		MimeType:     mimeType,
		Size:         written,
		Width:        width,
		Height:       height,
		PolicyID:     policy.ID,
		GroupID:      groupID,
		Variant:      "original",
		UploadedBy:   userID,
	}
	if err := s.svc.DB.WithContext(ctx).Create(&a).Error; err != nil {
		_ = os.Remove(diskPath)
		klog.ErrorS(err, "Upload: save db record")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	return &v1.AttachmentItem{
		ID:           a.ID,
		Name:         a.Name,
		OriginalName: a.OriginalName,
		URL:          a.URL,
		Type:         string(a.Type),
		MimeType:     a.MimeType,
		Size:         a.Size,
		Width:        a.Width,
		Height:       a.Height,
		Variant:      a.Variant,
		GroupID:      a.GroupID,
		CreatedAt:    a.CreatedAt,
	}, http.StatusOK, nil
}

// List returns paginated attachments with optional filters.
func (s *Service) List(ctx context.Context, q *v1.AttachmentListQuery) (*v1.AttachmentListResponse, int, error) {
	if q == nil {
		q = &v1.AttachmentListQuery{}
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	baseQ := s.svc.DB.WithContext(ctx).Model(&models.Attachment{})
	if q.RootsOnly == nil || (q.RootsOnly != nil && *q.RootsOnly) {
		baseQ = baseQ.Where("parent_id IS NULL")
	}
	if strings.TrimSpace(q.Type) != "" {
		baseQ = baseQ.Where("type = ?", models.AttachmentType(q.Type))
	}
	if kw := strings.TrimSpace(q.Keyword); kw != "" {
		pat := "%" + kw + "%"
		baseQ = baseQ.Where("(name ILIKE ? OR original_name ILIKE ?)", pat, pat)
	}
	if q.GroupID != nil && *q.GroupID > 0 {
		baseQ = baseQ.Where("group_id = ?", *q.GroupID)
	}

	var total int64
	if err := baseQ.Count(&total).Error; err != nil {
		klog.ErrorS(err, "List: count")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	listQ := s.svc.DB.WithContext(ctx).Model(&models.Attachment{})
	if q.RootsOnly == nil || (q.RootsOnly != nil && *q.RootsOnly) {
		listQ = listQ.Where("parent_id IS NULL")
	}
	if strings.TrimSpace(q.Type) != "" {
		listQ = listQ.Where("type = ?", models.AttachmentType(q.Type))
	}
	if kw := strings.TrimSpace(q.Keyword); kw != "" {
		pat := "%" + kw + "%"
		listQ = listQ.Where("(name ILIKE ? OR original_name ILIKE ?)", pat, pat)
	}
	if q.GroupID != nil && *q.GroupID > 0 {
		listQ = listQ.Where("group_id = ?", *q.GroupID)
	}

	var rows []models.Attachment
	if err := listQ.Preload("Group").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		klog.ErrorS(err, "List: find")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	// 根行展示：家族（根 + 子附件）在 article_attachments 中涉及的 distinct article_id 数量
	familyDistinctArticles := map[int64]int64{}
	if len(rows) > 0 {
		rootIDs := make([]int64, 0, len(rows))
		for i := range rows {
			rootIDs = append(rootIDs, rows[i].ID)
		}
		attToRoot := make(map[int64]int64, len(rootIDs)*2)
		for _, rid := range rootIDs {
			attToRoot[rid] = rid
		}
		var childRows []struct {
			ID       int64 `gorm:"column:id"`
			ParentID int64 `gorm:"column:parent_id"`
		}
		if err := s.svc.DB.WithContext(ctx).Model(&models.Attachment{}).
			Select("id", "parent_id").
			Where("parent_id IN ?", rootIDs).
			Scan(&childRows).Error; err != nil {
			klog.ErrorS(err, "List: children for ref count")
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		for i := range childRows {
			attToRoot[childRows[i].ID] = childRows[i].ParentID
		}
		allFamilyAttIDs := make([]int64, 0, len(attToRoot))
		for aid := range attToRoot {
			allFamilyAttIDs = append(allFamilyAttIDs, aid)
		}
		type pairRow struct {
			AttachmentID int64 `gorm:"column:attachment_id"`
			ArticleID    int64 `gorm:"column:article_id"`
		}
		var pairs []pairRow
		if len(allFamilyAttIDs) > 0 {
			if err := s.svc.DB.WithContext(ctx).Model(&models.ArticleAttachment{}).
				Select("attachment_id", "article_id").
				Where("attachment_id IN ?", allFamilyAttIDs).
				Scan(&pairs).Error; err != nil {
				klog.ErrorS(err, "List: article_attachments for family ref")
				return nil, http.StatusInternalServerError, errors.New("system error")
			}
		}
		perRootArts := make(map[int64]map[int64]struct{}, len(rootIDs))
		for _, rid := range rootIDs {
			perRootArts[rid] = make(map[int64]struct{})
		}
		for _, p := range pairs {
			rootID, ok := attToRoot[p.AttachmentID]
			if !ok {
				continue
			}
			if _, ok := perRootArts[rootID]; ok {
				perRootArts[rootID][p.ArticleID] = struct{}{}
			}
		}
		for rid, set := range perRootArts {
			familyDistinctArticles[rid] = int64(len(set))
		}
	}

	items := make([]v1.AttachmentItem, 0, len(rows))
	for i := range rows {
		a := &rows[i]
		gn := ""
		if a.Group != nil {
			gn = a.Group.Name
		}
		items = append(items, v1.AttachmentItem{
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
			RefCount:     familyDistinctArticles[a.ID],
			CreatedAt:    a.CreatedAt,
		})
	}
	return &v1.AttachmentListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK, nil
}

// Delete removes an attachment; deleting a root removes all descendants and their files.
func (s *Service) Delete(ctx context.Context, attachmentID int64) (*v1.DeleteAttachmentResponse, int, error) {
	if attachmentID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid attachment id")
	}
	var a models.Attachment
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", attachmentID).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("attachment not found")
		}
		klog.ErrorS(err, "Delete: load", "id", attachmentID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	if a.ParentID != nil {
		if err := s.deleteAttachmentLeaf(ctx, &a); err != nil {
			klog.ErrorS(err, "Delete: leaf", "id", attachmentID)
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		return &v1.DeleteAttachmentResponse{ID: attachmentID}, http.StatusOK, nil
	}

	var childIDs []int64
	if err := s.svc.DB.WithContext(ctx).Model(&models.Attachment{}).
		Where("parent_id = ?", a.ID).Pluck("id", &childIDs).Error; err != nil {
		klog.ErrorS(err, "Delete: list children", "id", attachmentID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	for _, cid := range childIDs {
		var c models.Attachment
		if err := s.svc.DB.WithContext(ctx).Where("id = ?", cid).First(&c).Error; err != nil {
			continue
		}
		if err := s.deleteAttachmentLeaf(ctx, &c); err != nil {
			klog.ErrorS(err, "Delete: child", "id", cid)
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}
	if err := s.deleteAttachmentLeaf(ctx, &a); err != nil {
		klog.ErrorS(err, "Delete: root", "id", attachmentID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return &v1.DeleteAttachmentResponse{ID: attachmentID}, http.StatusOK, nil
}

func (s *Service) deleteAttachmentLeaf(ctx context.Context, a *models.Attachment) error {
	if err := s.svc.DB.WithContext(ctx).Where("attachment_id = ?", a.ID).Delete(&models.ArticleAttachment{}).Error; err != nil {
		return err
	}
	if err := s.svc.DB.WithContext(ctx).Delete(a).Error; err != nil {
		return err
	}
	if a.Path != "" {
		if err := os.Remove(a.Path); err != nil && !os.IsNotExist(err) {
			klog.ErrorS(err, "deleteAttachmentLeaf: remove file", "path", a.Path)
		}
	}
	return nil
}
