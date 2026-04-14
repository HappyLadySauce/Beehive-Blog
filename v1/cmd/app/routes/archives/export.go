package archives

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/markdownfrontmatter"
	"gorm.io/gorm"
)

const (
	exportFormatMarkdown = "markdown"
	exportFormatHTML     = "html"
	exportFormatPDF      = "pdf"
)

// ExportArticle 导出文章为指定格式。
// 返回值：(内容字节, Content-Type, HTTP 状态码, error)
// 支持 markdown（原始 Markdown）和 html（简单 HTML 包装）；pdf 返回 501。
func (a *ArticleAdmin) ExportArticle(ctx context.Context, articleID int64, format string) ([]byte, string, int, error) {
	if articleID <= 0 {
		return nil, "", http.StatusBadRequest, errors.New("invalid article id")
	}
	if format == "" {
		format = exportFormatMarkdown
	}
	if format != exportFormatMarkdown && format != exportFormatHTML && format != exportFormatPDF {
		return nil, "", http.StatusBadRequest, fmt.Errorf("unsupported format %q; supported: markdown, html, pdf", format)
	}
	if format == exportFormatPDF {
		return nil, "", http.StatusNotImplemented, errors.New("pdf export is not implemented")
	}

	var art models.Article
	if err := a.svc.DB.WithContext(ctx).
		Preload("Author").Preload("Category").Preload("Tags").
		Where("id = ? AND deleted_at IS NULL", articleID).First(&art).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", http.StatusNotFound, errors.New("article not found")
		}
		return nil, "", http.StatusInternalServerError, errors.New("system error")
	}

	switch format {
	case exportFormatMarkdown:
		data := buildMarkdownExport(&art)
		return data, "text/markdown; charset=utf-8", http.StatusOK, nil
	case exportFormatHTML:
		data := buildHTMLExport(&art)
		return data, "text/html; charset=utf-8", http.StatusOK, nil
	}
	return nil, "", http.StatusBadRequest, errors.New("unsupported format")
}

// buildMarkdownExport 构建带 Front Matter 的 Markdown 导出内容。
func buildMarkdownExport(art *models.Article) []byte {
	var sb []byte
	sb = append(sb, []byte("---\n")...)
	sb = append(sb, []byte(fmt.Sprintf("title: %q\n", art.Title))...)
	sb = append(sb, []byte(fmt.Sprintf("slug: %q\n", art.Slug))...)
	if art.PublishedAt != nil {
		sb = append(sb, []byte(fmt.Sprintf("date: %s\n", art.PublishedAt.Format("2006-01-02T15:04:05Z07:00")))...)
	}
	if art.Category != nil {
		sb = append(sb, []byte(fmt.Sprintf("categories: [%q]\n", art.Category.Name))...)
	}
	if len(art.Tags) > 0 {
		sb = append(sb, []byte("tags: [")...)
		for i, t := range art.Tags {
			if i > 0 {
				sb = append(sb, ',')
			}
			sb = append(sb, []byte(fmt.Sprintf("%q", t.Name))...)
		}
		sb = append(sb, []byte("]\n")...)
	}
	sb = append(sb, []byte(fmt.Sprintf("status: %q\n", art.Status))...)
	sb = append(sb, []byte("---\n\n")...)
	contentBody := art.Content
	if _, b, ok := markdownfrontmatter.SplitFrontMatter(art.Content); ok {
		contentBody = b
	}
	sb = append(sb, []byte(contentBody)...)
	return sb
}

// buildHTMLExport 将 Markdown 内容包装为简单 HTML 页面（不引入外部渲染库）。
// 正文以 <pre> 标签包裹原始 Markdown，适合在浏览器中查看或二次处理。
func buildHTMLExport(art *models.Article) []byte {
	escapedTitle := html.EscapeString(art.Title)
	contentBody := art.Content
	if _, b, ok := markdownfrontmatter.SplitFrontMatter(art.Content); ok {
		contentBody = b
	}
	escapedContent := html.EscapeString(contentBody)
	body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
  <style>
    body { font-family: sans-serif; max-width: 900px; margin: 40px auto; padding: 0 20px; }
    pre { white-space: pre-wrap; word-break: break-word; background: #f6f8fa; padding: 16px; border-radius: 6px; }
  </style>
</head>
<body>
  <h1>%s</h1>
  <pre>%s</pre>
</body>
</html>`, escapedTitle, escapedTitle, escapedContent)
	return []byte(body)
}
