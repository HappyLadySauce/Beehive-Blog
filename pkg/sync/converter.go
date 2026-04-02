package sync

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"gopkg.in/yaml.v3"
)

// HexoFrontMatter 与 hexo-sync-design 对齐，不含明文密码。
type HexoFrontMatter struct {
	Title            string    `yaml:"title"`
	Slug             string    `yaml:"slug,omitempty"`
	Description      string    `yaml:"description,omitempty"`
	Date             time.Time `yaml:"date"`
	Updated          time.Time `yaml:"updated,omitempty"`
	Categories       []string  `yaml:"categories,omitempty"`
	Tags             []string  `yaml:"tags,omitempty"`
	Cover            string    `yaml:"cover,omitempty"`
	Pin              bool      `yaml:"pin,omitempty"`
	PinOrder         int       `yaml:"pin_order,omitempty"`
	Views            int64     `yaml:"views,omitempty"`
	BeehiveID        int64     `yaml:"beehive_id"`
	BeehiveProtected bool      `yaml:"beehive_protected,omitempty"`
}

var slugSafeReplacer = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// ArticleToHexoMarkdown 将文章转为 Hexo 可用的 Markdown 文件内容（UTF-8）。
func ArticleToHexoMarkdown(article *models.Article, categoryName string, tagNames []string) ([]byte, error) {
	if article == nil {
		return nil, fmt.Errorf("article is nil")
	}

	postDate := time.Now()
	if article.PublishedAt != nil {
		postDate = *article.PublishedAt
	} else if !article.CreatedAt.IsZero() {
		postDate = article.CreatedAt
	}

	updated := article.UpdatedAt
	if updated.IsZero() {
		updated = postDate
	}

	slugForURL := strings.TrimSpace(article.Slug)
	if slugForURL == "" {
		slugForURL = sanitizeFileSlug("", article.ID)
	} else {
		slugForURL = sanitizeFileSlug(article.Slug, article.ID)
	}

	fm := HexoFrontMatter{
		Title:       article.Title,
		Slug:        slugForURL,
		Description: article.Summary,
		Date:        postDate,
		Updated:     updated,
		Cover:       article.CoverImage,
		Pin:         article.IsPinned,
		PinOrder:    article.PinOrder,
		Views:       article.ViewCount,
		BeehiveID:   article.ID,
	}
	if article.Password != "" {
		fm.BeehiveProtected = true
	}

	if strings.TrimSpace(categoryName) != "" {
		fm.Categories = []string{categoryName}
	}

	if len(tagNames) > 0 {
		fm.Tags = append([]string(nil), tagNames...)
	}

	var ymlBody bytes.Buffer
	enc := yaml.NewEncoder(&ymlBody)
	enc.SetIndent(2)
	if err := enc.Encode(&fm); err != nil {
		return nil, fmt.Errorf("encode front matter: %w", err)
	}
	_ = enc.Close()

	var out bytes.Buffer
	out.WriteString("---\n")
	out.Write(ymlBody.Bytes())
	out.WriteString("---\n\n")
	out.WriteString(article.Content)
	if len(article.Content) == 0 || !strings.HasSuffix(article.Content, "\n") {
		out.WriteByte('\n')
	}
	return out.Bytes(), nil
}

// GenerateHexoFileName 生成 beehive-{id}-{slug}.md，slug 经文件系统安全化。
func GenerateHexoFileName(article *models.Article) string {
	if article == nil {
		return ""
	}
	slugPart := sanitizeFileSlug(article.Slug, article.ID)
	return fmt.Sprintf("beehive-%d-%s.md", article.ID, slugPart)
}

func sanitizeFileSlug(slug string, articleID int64) string {
	s := strings.TrimSpace(slug)
	if s == "" {
		return fmt.Sprintf("post-%d", articleID)
	}
	s = slugSafeReplacer.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return fmt.Sprintf("post-%d", articleID)
	}
	return out
}

// SortedTagNames 从关联标签提取名称并排序，保证生成稳定。
func SortedTagNames(tags []models.Tag) []string {
	if len(tags) == 0 {
		return nil
	}
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		n := strings.TrimSpace(t.Name)
		if n != "" {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names
}

// CategoryNameOrEmpty 返回用于写入 front-matter 的分类显示名。
func CategoryNameOrEmpty(c *models.Category) string {
	if c == nil {
		return ""
	}
	return strings.TrimSpace(c.Name)
}
