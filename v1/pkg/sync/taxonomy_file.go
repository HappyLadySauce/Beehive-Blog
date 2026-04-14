package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/v1/pkg/utils/color"
)

const defaultHexoTagColor = "#3B82F6"

// BeehiveTaxonomyDoc 写入 Hexo source/_data/beehive_taxonomy.json，供主题与 beehive-taxonomy-merge 脚本使用。
type BeehiveTaxonomyDoc struct {
	TagMap      map[string]string `json:"tag_map"`
	CategoryMap map[string]string `json:"category_map"`
	TagColors   map[string]string `json:"tag_colors"`
}

type tagTaxonomyRow struct {
	Name  string `gorm:"column:name"`
	Slug  string `gorm:"column:slug"`
	Color string `gorm:"column:color"`
}

type categoryTaxonomyRow struct {
	Name string `gorm:"column:name"`
	Slug string `gorm:"column:slug"`
}

// WriteBeehiveTaxonomyJSON 从数据库拉取标签与分类，写入 {hexoRoot}/source/_data/beehive_taxonomy.json。
// 与单篇文章无关；在 SyncAll 结束或标签/分类变更后调用。内容未变化时跳过写盘。
func (s *SyncService) WriteBeehiveTaxonomyJSON(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("sync service or db is nil")
	}
	doc, err := s.buildBeehiveTaxonomyDoc(ctx)
	if err != nil {
		return err
	}
	payload, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')

	path := taxonomyJSONPath(s.generateWorkdir)
	if prev, err := os.ReadFile(path); err == nil && bytes.Equal(prev, payload) {
		return nil
	}
	return atomicWriteFile(path, payload, 0o644)
}

func taxonomyJSONPath(hexoRoot string) string {
	return filepath.Join(hexoRoot, "source", "_data", "beehive_taxonomy.json")
}

func (s *SyncService) buildBeehiveTaxonomyDoc(ctx context.Context) (*BeehiveTaxonomyDoc, error) {
	var tagRows []tagTaxonomyRow
	if err := s.db.WithContext(ctx).Model(&models.Tag{}).
		Select("name", "slug", "color").
		Order("id ASC").
		Find(&tagRows).Error; err != nil {
		return nil, err
	}

	var catRows []categoryTaxonomyRow
	if err := s.db.WithContext(ctx).Model(&models.Category{}).
		Select("name", "slug").
		Order("id ASC").
		Find(&catRows).Error; err != nil {
		return nil, err
	}

	out := &BeehiveTaxonomyDoc{
		TagMap:      make(map[string]string),
		CategoryMap: make(map[string]string),
		TagColors:   make(map[string]string),
	}

	for _, r := range tagRows {
		name := strings.TrimSpace(r.Name)
		slug := strings.TrimSpace(r.Slug)
		if name == "" || slug == "" {
			continue
		}
		out.TagMap[name] = slug
		col := strings.TrimSpace(r.Color)
		if h := color.NormalizeToHex(col); h != "" {
			out.TagColors[slug] = h
		} else if color.ValidHex(col) {
			out.TagColors[slug] = strings.ToUpper(col)
		} else {
			out.TagColors[slug] = defaultHexoTagColor
		}
	}

	for _, r := range catRows {
		name := strings.TrimSpace(r.Name)
		slug := strings.TrimSpace(r.Slug)
		if name == "" || slug == "" {
			continue
		}
		out.CategoryMap[name] = slug
	}

	return out, nil
}

// TaxonomyFilePath 返回 beehive_taxonomy.json 的绝对路径（用于测试或诊断）。
func (s *SyncService) TaxonomyFilePath() string {
	if s == nil {
		return ""
	}
	return taxonomyJSONPath(s.generateWorkdir)
}
