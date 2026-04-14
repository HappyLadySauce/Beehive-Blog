package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
)

// SyncAllPages 全量同步已发布页面并清理 beehive-pages 下孤儿目录。
func (s *SyncService) SyncAllPages(ctx context.Context) (total, created, updated, deleted int, files []string, err error) {
	var pages []models.Page
	if err := s.db.WithContext(ctx).
		Where("status = ?", models.ArticleStatusPublished).
		Order("id ASC").
		Find(&pages).Error; err != nil {
		return 0, 0, 0, 0, nil, err
	}

	active := make(map[int64]struct{}, len(pages))
	files = make([]string, 0, len(pages))
	for i := range pages {
		p := &pages[i]
		active[p.ID] = struct{}{}
		action, e := s.writePageFile(ctx, p)
		if e != nil {
			return total, created, updated, deleted, files, fmt.Errorf("sync page %d: %w", p.ID, e)
		}
		total++
		switch action {
		case SyncActionCreate:
			created++
		case SyncActionUpdate:
			updated++
		}
		files = append(files, GenerateBeehivePageDirName(p))
	}

	n, e := s.cleanupOrphanedPages(ctx, active)
	if e != nil {
		return total, created, updated, deleted, files, e
	}
	deleted = n
	return total, created, updated, deleted, files, nil
}

// SyncSinglePage 按 ID 同步单页：已发布则写入，否则删除受管目录。
func (s *SyncService) SyncSinglePage(ctx context.Context, pageID int64) error {
	var p models.Page
	if err := s.db.WithContext(ctx).First(&p, pageID).Error; err != nil {
		return err
	}
	if p.Status != models.ArticleStatusPublished {
		return s.DeletePageFile(&p)
	}
	if _, err := s.writePageFile(ctx, &p); err != nil {
		return err
	}
	return s.writeLastSyncMarker(ctx)
}

// DeletePageFile 删除该页面 ID 对应的全部 beehive-{id}-* 目录。
func (s *SyncService) DeletePageFile(page *models.Page) error {
	if page == nil {
		return fmt.Errorf("page is nil")
	}
	return s.removePageDirsForPageID(page.ID)
}

func (s *SyncService) writePageFile(ctx context.Context, page *models.Page) (SyncAction, error) {
	gen := *page
	content, err := s.rewriteForHexo(ctx, page.Content)
	if err != nil {
		return "", err
	}
	gen.Content = content
	payload, err := PageToHexoMarkdown(&gen)
	if err != nil {
		return "", err
	}
	dirName := GenerateBeehivePageDirName(page)
	fullDir := filepath.Join(s.beehivePagesDirAbs, dirName)
	indexPath := filepath.Join(fullDir, "index.md")

	had, err := s.hasBeehivePageDirForPageID(page.ID)
	if err != nil {
		return "", err
	}
	if err := s.removePageDirsForPageID(page.ID); err != nil {
		return "", err
	}
	if err := atomicWriteFile(indexPath, payload, 0o644); err != nil {
		return "", err
	}
	if had {
		return SyncActionUpdate, nil
	}
	return SyncActionCreate, nil
}

func (s *SyncService) hasBeehivePageDirForPageID(id int64) (bool, error) {
	entries, err := os.ReadDir(s.beehivePagesDirAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "beehive-") {
			continue
		}
		pid, ok := parseBeehiveManagedDirPrefixID(name)
		if ok && pid == id {
			return true, nil
		}
	}
	return false, nil
}

func (s *SyncService) removePageDirsForPageID(id int64) error {
	if err := os.MkdirAll(s.beehivePagesDirAbs, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(s.beehivePagesDirAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(s.beehivePagesDirAbs, 0o755)
		}
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "beehive-") {
			continue
		}
		pid, ok := parseBeehiveManagedDirPrefixID(name)
		if !ok || pid != id {
			continue
		}
		_ = os.RemoveAll(filepath.Join(s.beehivePagesDirAbs, name))
	}
	return nil
}

// parseBeehiveManagedDirPrefixID 从 beehive-{id}-{slug} 目录或文件名前缀解析 id。
func parseBeehiveManagedDirPrefixID(name string) (int64, bool) {
	return parseBeehivePostFileID(name + ".md")
}

func (s *SyncService) cleanupOrphanedPages(ctx context.Context, active map[int64]struct{}) (int, error) {
	_ = ctx
	entries, err := os.ReadDir(s.beehivePagesDirAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, os.MkdirAll(s.beehivePagesDirAbs, 0o755)
		}
		return 0, err
	}
	removed := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "beehive-") {
			continue
		}
		id, ok := parseBeehiveManagedDirPrefixID(name)
		if !ok {
			continue
		}
		if _, keep := active[id]; keep {
			continue
		}
		if err := os.RemoveAll(filepath.Join(s.beehivePagesDirAbs, name)); err != nil && !os.IsNotExist(err) {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

// PublishedPageCount 返回已发布且未软删的页面数。
func (s *SyncService) PublishedPageCount(ctx context.Context) (int64, error) {
	var c int64
	err := s.db.WithContext(ctx).Model(&models.Page{}).
		Where("status = ?", models.ArticleStatusPublished).
		Count(&c).Error
	return c, err
}

// LocalBeehivePageCount 统计 beehive-pages 下受管目录（含 index.md）数量。
func (s *SyncService) LocalBeehivePageCount() (int, error) {
	entries, err := os.ReadDir(s.beehivePagesDirAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "beehive-") {
			continue
		}
		idx := filepath.Join(s.beehivePagesDirAbs, name, "index.md")
		if st, err := os.Stat(idx); err == nil && !st.IsDir() {
			n++
		}
	}
	return n, nil
}
