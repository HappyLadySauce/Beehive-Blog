package sync

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"gopkg.in/yaml.v3"
)

// HexoPageFrontMatter 独立页面 front-matter（不出现在文章列表聚合中）。
type HexoPageFrontMatter struct {
	Title         string    `yaml:"title"`
	Layout        string    `yaml:"layout"`
	Permalink     string    `yaml:"permalink"`
	Date          time.Time `yaml:"date"`
	Updated       time.Time `yaml:"updated,omitempty"`
	BeehivePageID int64     `yaml:"beehive_page_id"`
	Views         int64     `yaml:"views,omitempty"`
}

// PageToHexoMarkdown 将 Page 转为 Hexo 独立页面 Markdown（目录内 index.md 内容）。
func PageToHexoMarkdown(page *models.Page) ([]byte, error) {
	if page == nil {
		return nil, fmt.Errorf("page is nil")
	}

	pageDate := page.CreatedAt
	if pageDate.IsZero() {
		pageDate = time.Now()
	}
	updated := page.UpdatedAt
	if updated.IsZero() {
		updated = pageDate
	}

	permalinkSeg := strings.TrimSpace(page.Slug)
	if permalinkSeg == "" {
		permalinkSeg = sanitizeFileSlug("", page.ID)
	} else {
		permalinkSeg = sanitizeFileSlug(page.Slug, page.ID)
	}

	fm := HexoPageFrontMatter{
		Title:         page.Title,
		Layout:        "page",
		Permalink:     permalinkSeg + "/",
		Date:          pageDate,
		Updated:       updated,
		BeehivePageID: page.ID,
		Views:         page.ViewCount,
	}

	var ymlBody bytes.Buffer
	enc := yaml.NewEncoder(&ymlBody)
	enc.SetIndent(2)
	if err := enc.Encode(&fm); err != nil {
		return nil, fmt.Errorf("encode page front matter: %w", err)
	}
	_ = enc.Close()

	var out bytes.Buffer
	out.WriteString("---\n")
	out.Write(ymlBody.Bytes())
	out.WriteString("---\n\n")
	out.WriteString(page.Content)
	if len(page.Content) == 0 || !strings.HasSuffix(page.Content, "\n") {
		out.WriteByte('\n')
	}
	return out.Bytes(), nil
}

// GenerateBeehivePageDirName 生成 source/beehive-pages/ 下的目录名：beehive-{id}-{slug}。
func GenerateBeehivePageDirName(page *models.Page) string {
	if page == nil {
		return ""
	}
	slugPart := sanitizeFileSlug(page.Slug, page.ID)
	return fmt.Sprintf("beehive-%d-%s", page.ID, slugPart)
}
