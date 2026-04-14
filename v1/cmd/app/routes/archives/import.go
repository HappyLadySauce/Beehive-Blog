package archives

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/markdownfrontmatter"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/color"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/slug"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

const (
	maxImportFiles           = 50
	maxImportMDBytes         = 4 << 20 // 单篇 Markdown
	maxImportZipBytes        = 32 << 20
	maxImportZipUncompressed = 20 << 20
)

type importSourceFile struct {
	name string
	data []byte
}

// ImportMarkdownUpload 处理 multipart：files（多 .md）与/或 archive（.zip）。
func (a *ArticleAdmin) ImportMarkdownUpload(ctx context.Context, adminUserID int64, c *gin.Context) (*v1.ImportArticlesResponse, int, error) {
	if err := c.Request.ParseMultipartForm(maxImportZipBytes); err != nil {
		return nil, http.StatusBadRequest, err
	}
	form := c.Request.MultipartForm
	if form == nil {
		return nil, http.StatusBadRequest, errors.New("no multipart form")
	}
	defer func() { _ = form.RemoveAll() }()

	var sources []importSourceFile

	for _, fh := range form.File["files"] {
		if fh.Size > maxImportMDBytes {
			return nil, http.StatusRequestEntityTooLarge, errors.New("markdown file too large")
		}
		src, err := fh.Open()
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		data, err := io.ReadAll(io.LimitReader(src, maxImportMDBytes+1))
		_ = src.Close()
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		if len(data) > maxImportMDBytes {
			return nil, http.StatusRequestEntityTooLarge, errors.New("markdown file too large")
		}
		if len(data) == 0 {
			continue
		}
		sources = append(sources, importSourceFile{name: fh.Filename, data: data})
	}

	archives := form.File["archive"]
	if len(archives) > 1 {
		return nil, http.StatusBadRequest, errors.New("only one zip archive allowed")
	}
	if len(archives) == 1 {
		ah := archives[0]
		if ah.Size > maxImportZipBytes {
			return nil, http.StatusRequestEntityTooLarge, errors.New("archive too large")
		}
		rc, err := ah.Open()
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		zdata, err := io.ReadAll(io.LimitReader(rc, maxImportZipBytes+1))
		_ = rc.Close()
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		if len(zdata) > maxImportZipBytes {
			return nil, http.StatusRequestEntityTooLarge, errors.New("archive too large")
		}
		zipped, err := extractMDFromZip(zdata)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		sources = append(sources, zipped...)
	}

	if len(sources) == 0 {
		return nil, http.StatusBadRequest, errors.New("no markdown files")
	}
	if len(sources) > maxImportFiles {
		return nil, http.StatusBadRequest, fmt.Errorf("at most %d files per import", maxImportFiles)
	}

	out := &v1.ImportArticlesResponse{
		Items:  make([]v1.ImportArticleCreatedItem, 0),
		Errors: make([]v1.ImportArticleErrorItem, 0),
	}

	for _, sf := range sources {
		item, errMsg := a.importOneMarkdown(ctx, adminUserID, sf.name, sf.data)
		if errMsg != "" {
			out.Errors = append(out.Errors, v1.ImportArticleErrorItem{File: sf.name, Reason: errMsg})
			continue
		}
		out.Created++
		out.Items = append(out.Items, *item)
	}

	return out, http.StatusOK, nil
}

func extractMDFromZip(zipData []byte) ([]importSourceFile, error) {
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, err
	}
	var totalUncomp uint64
	var out []importSourceFile
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !isSafeZipEntryPath(f.Name) {
			continue
		}
		lower := strings.ToLower(f.Name)
		if !strings.HasSuffix(lower, ".md") && !strings.HasSuffix(lower, ".markdown") {
			continue
		}
		if f.UncompressedSize64 > uint64(maxImportMDBytes) {
			return nil, fmt.Errorf("zip entry too large: %s", f.Name)
		}
		totalUncomp += f.UncompressedSize64
		if totalUncomp > uint64(maxImportZipUncompressed) {
			return nil, errors.New("zip uncompressed size too large")
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(rc, maxImportMDBytes+1))
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
		if len(data) > maxImportMDBytes {
			return nil, fmt.Errorf("zip entry too large: %s", f.Name)
		}
		out = append(out, importSourceFile{name: filepath.Base(f.Name), data: data})
		if len(out) > maxImportFiles {
			return nil, errors.New("too many markdown files in zip")
		}
	}
	if len(out) == 0 {
		return nil, errors.New("no markdown entries in zip")
	}
	return out, nil
}

func isSafeZipEntryPath(name string) bool {
	name = filepath.ToSlash(strings.TrimSpace(name))
	if name == "" || strings.Contains(name, "..") {
		return false
	}
	if strings.HasPrefix(name, "/") {
		return false
	}
	for _, seg := range strings.Split(name, "/") {
		if seg == ".." {
			return false
		}
	}
	return true
}

func (a *ArticleAdmin) importOneMarkdown(ctx context.Context, adminUserID int64, filename string, data []byte) (*v1.ImportArticleCreatedItem, string) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	yamlStr, body, ok := markdownfrontmatter.SplitFrontMatter(text)
	if !ok {
		return nil, "missing or invalid yaml front matter"
	}
	var fm map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &fm); err != nil {
		return nil, "invalid front matter yaml"
	}
	title, _ := fm["title"].(string)
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, "title is required in front matter"
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, "content is empty"
	}

	slugStr := strings.TrimSpace(getString(fm["slug"]))
	if slugStr != "" {
		var normOK bool
		slugStr, normOK = slug.Normalize(slugStr)
		if !normOK {
			slugStr = ""
		} else {
			taken, err := a.slugTaken(ctx, slugStr, 0)
			if err != nil {
				klog.ErrorS(err, "import slug check")
				return nil, "system error"
			}
			if taken {
				slugStr = ""
			}
		}
	}

	summary := strings.TrimSpace(getString(fm["summary"]))

	catHint := firstCategoryString(fm["categories"])
	categoryID, err := a.resolveCategoryIDFromImport(ctx, catHint)
	if err != nil {
		klog.ErrorS(err, "import category")
		return nil, "system error"
	}

	tagNames := stringSliceFromYAML(fm["tags"])
	tagIDs := make([]int64, 0, len(tagNames))
	for _, tn := range tagNames {
		tn = strings.TrimSpace(tn)
		if tn == "" {
			continue
		}
		tid, err := a.findOrCreateTagByName(ctx, tn)
		if err != nil {
			return nil, err.Error()
		}
		tagIDs = append(tagIDs, tid)
	}
	tagIDs = normalizeTagIDs(tagIDs)

	statusStr := strings.TrimSpace(getString(fm["status"]))
	if statusStr == "" {
		statusStr = string(models.ArticleStatusDraft)
	}
	if !isValidArticleStatus(statusStr) {
		return nil, "invalid status in front matter"
	}

	var publishedAt *string
	dateStr := strings.TrimSpace(getString(fm["date"]))
	if dateStr != "" {
		norm := normalizeImportDateString(dateStr)
		publishedAt = &norm
	}

	req := &v1.CreateArticleRequest{
		Title:       title,
		Slug:        slugStr,
		Content:     body,
		Summary:     summary,
		CategoryID:  categoryID,
		TagIDs:      tagIDs,
		Status:      statusStr,
		PublishedAt: publishedAt,
	}

	detail, code, err := a.CreateArticle(ctx, adminUserID, req, nil)
	if err != nil {
		if code == http.StatusConflict {
			return nil, "slug conflict"
		}
		if code == http.StatusBadRequest {
			return nil, err.Error()
		}
		klog.ErrorS(err, "import create article", "file", filename)
		return nil, "system error"
	}
	return &v1.ImportArticleCreatedItem{ID: detail.ID, Title: detail.Title}, ""
}

func normalizeImportDateString(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC().Format(time.RFC3339)
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.UTC); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.UTC); err == nil {
		return t.UTC().Format(time.RFC3339)
	}
	return s
}

func getString(v interface{}) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func firstCategoryString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case []interface{}:
		for _, it := range x {
			if s, ok := it.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func stringSliceFromYAML(v interface{}) []string {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return nil
		}
		return []string{s}
	case []interface{}:
		var out []string
		for _, it := range x {
			if s, ok := it.(string); ok {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	default:
		return nil
	}
}

func isValidArticleStatus(s string) bool {
	switch models.ArticleStatus(s) {
	case models.ArticleStatusDraft, models.ArticleStatusPublished, models.ArticleStatusArchived,
		models.ArticleStatusPrivate, models.ArticleStatusScheduled:
		return true
	default:
		return false
	}
}

func (a *ArticleAdmin) resolveCategoryIDFromImport(ctx context.Context, catHint string) (*int64, error) {
	catHint = strings.TrimSpace(catHint)
	if catHint == "" {
		return a.resolveCategoryIDForCreate(ctx, nil)
	}
	var c models.Category
	err := a.svc.DB.WithContext(ctx).
		Where("LOWER(name) = LOWER(?) OR LOWER(slug) = LOWER(?)", catHint, catHint).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return a.resolveCategoryIDForCreate(ctx, nil)
	}
	if err != nil {
		return nil, err
	}
	cid := c.ID
	return &cid, nil
}

func (a *ArticleAdmin) importTagSlugTaken(ctx context.Context, sl string, excludeID int64) (bool, error) {
	q := a.svc.DB.WithContext(ctx).Model(&models.Tag{}).Where("slug = ?", sl)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (a *ArticleAdmin) importAllocUniqueTagSlug(ctx context.Context, base string, excludeID int64) (string, error) {
	if base == "" {
		base = slug.Fallback(time.Now().UnixNano())
	}
	for i := 0; i < 50; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i+1)
		}
		taken, err := a.importTagSlugTaken(ctx, candidate, excludeID)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}
	return "", errors.New("could not allocate unique tag slug")
}

func (a *ArticleAdmin) findOrCreateTagByName(ctx context.Context, name string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("empty tag name")
	}
	var existing models.Tag
	err := a.svc.DB.WithContext(ctx).Where("LOWER(name) = LOWER(?)", name).First(&existing).Error
	if err == nil {
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	col := color.RandomHexColor()
	if h := color.NormalizeToHex(col); h != "" {
		col = h
	}
	slugBase := slug.FromTitle(name)
	slugStr, err := a.importAllocUniqueTagSlug(ctx, slugBase, 0)
	if err != nil {
		return 0, err
	}
	t := &models.Tag{
		Name:      name,
		Slug:      slugStr,
		Color:     col,
		SortOrder: 0,
	}
	if err := a.svc.DB.WithContext(ctx).Create(t).Error; err != nil {
		return 0, err
	}
	return t.ID, nil
}
