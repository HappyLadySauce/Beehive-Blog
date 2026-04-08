package archives

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
)

const (
	batchExportMaxIDs = 100
)

// BatchExportArticlesZIP 将多篇文章打包为 ZIP（markdown 或 html）。
// 返回：(zip 字节, Content-Type, HTTP 状态码, error)
func (a *ArticleAdmin) BatchExportArticlesZIP(ctx context.Context, ids []int64, format string) ([]byte, string, int, error) {
	if len(ids) == 0 {
		return nil, "", http.StatusBadRequest, errors.New("ids required")
	}
	if len(ids) > batchExportMaxIDs {
		return nil, "", http.StatusBadRequest, fmt.Errorf("at most %d articles per export", batchExportMaxIDs)
	}
	if format == "" {
		format = exportFormatMarkdown
	}
	if format != exportFormatMarkdown && format != exportFormatHTML {
		return nil, "", http.StatusBadRequest, errors.New("unsupported format")
	}

	var articles []models.Article
	if err := a.svc.DB.WithContext(ctx).
		Preload("Author").Preload("Category").Preload("Tags").
		Where("id IN ? AND deleted_at IS NULL", ids).
		Find(&articles).Error; err != nil {
		return nil, "", http.StatusInternalServerError, errors.New("system error")
	}
	if len(articles) == 0 {
		return nil, "", http.StatusBadRequest, errors.New("no articles found for given ids")
	}

	usedNames := make(map[string]struct{}, len(articles))
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for i := range articles {
		art := &articles[i]
		var data []byte
		var ext string
		switch format {
		case exportFormatMarkdown:
			data = buildMarkdownExport(art)
			ext = ".md"
		case exportFormatHTML:
			data = buildHTMLExport(art)
			ext = ".html"
		default:
			_ = zw.Close()
			return nil, "", http.StatusBadRequest, errors.New("unsupported format")
		}
		entryName := uniqueZipEntryName(art.Slug, ext, usedNames)
		w, err := zw.Create(entryName)
		if err != nil {
			_ = zw.Close()
			return nil, "", http.StatusInternalServerError, errors.New("system error")
		}
		if _, err := w.Write(data); err != nil {
			_ = zw.Close()
			return nil, "", http.StatusInternalServerError, errors.New("system error")
		}
	}
	if err := zw.Close(); err != nil {
		return nil, "", http.StatusInternalServerError, errors.New("system error")
	}
	return buf.Bytes(), "application/zip", http.StatusOK, nil
}

func uniqueZipEntryName(slug, ext string, used map[string]struct{}) string {
	base := sanitizeZipBaseName(slug)
	if base == "" {
		base = "article"
	}
	name := base + ext
	for i := 0; ; i++ {
		if _, ok := used[name]; !ok {
			used[name] = struct{}{}
			return name
		}
		name = fmt.Sprintf("%s-%d%s", base, i+1, ext)
	}
}

func sanitizeZipBaseName(slug string) string {
	s := strings.TrimSpace(slug)
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r < 32 || r == 127:
			b.WriteByte('_')
		case strings.ContainsRune(`\/:*?"<>|`, r):
			b.WriteByte('_')
		default:
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), ". ")
	if out == "" {
		return ""
	}
	// zip slip: no path components
	out = path.Base(out)
	if out == "." || out == "/" {
		return ""
	}
	return out
}
