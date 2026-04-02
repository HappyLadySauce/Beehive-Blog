package sync

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const redisLastSyncKey = "beehive:hexo_sync:last_sync_at"

// SyncService 将已发布文章写入 Hexo _posts 目录。
type SyncService struct {
	postsDirAbs     string
	generateWorkdir string
	generateArgs    []string
	db              *gorm.DB
	rdb             *redis.Client
}

// NewSyncService 基于配置解析绝对路径并构造同步服务。
func NewSyncService(postsDir, generateWorkdir string, generateArgs []string, db *gorm.DB, rdb *redis.Client) (*SyncService, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	postsAbs, err := filepath.Abs(postsDir)
	if err != nil {
		return nil, fmt.Errorf("resolve posts dir: %w", err)
	}
	genAbs, err := filepath.Abs(generateWorkdir)
	if err != nil {
		return nil, fmt.Errorf("resolve hexo workdir: %w", err)
	}
	ga := append([]string(nil), generateArgs...)
	return &SyncService{
		postsDirAbs:     postsAbs,
		generateWorkdir: genAbs,
		generateArgs:    ga,
		db:              db,
		rdb:             rdb,
	}, nil
}

// PostsDirAbs 返回解析后的 _posts 绝对路径。
func (s *SyncService) PostsDirAbs() string {
	return s.postsDirAbs
}

// SyncAll 全量同步已发布文章并清理孤儿 beehive-*.md 文件。
func (s *SyncService) SyncAll(ctx context.Context) (*SyncResult, error) {
	var articles []models.Article
	if err := s.db.WithContext(ctx).
		Where("status = ?", models.ArticleStatusPublished).
		Preload("Category").
		Preload("Tags").
		Order("published_at DESC NULLS LAST, id DESC").
		Find(&articles).Error; err != nil {
		return nil, err
	}

	activeIDs := make(map[int64]struct{}, len(articles))
	res := &SyncResult{Total: len(articles)}

	for i := range articles {
		a := &articles[i]
		activeIDs[a.ID] = struct{}{}
		action, err := s.writeArticleFile(ctx, a)
		if err != nil {
			return res, fmt.Errorf("sync article %d: %w", a.ID, err)
		}
		switch action {
		case SyncActionCreate:
			res.Created++
		case SyncActionUpdate:
			res.Updated++
		}
		res.Files = append(res.Files, GenerateHexoFileName(a))
	}

	n, err := s.cleanupOrphaned(ctx, activeIDs)
	if err != nil {
		return res, err
	}
	res.Deleted = n

	if s.rdb != nil {
		_ = s.rdb.Set(ctx, redisLastSyncKey, time.Now().UTC().Format(time.RFC3339Nano), 0).Err()
	}

	return res, nil
}

// SyncSingle 按 ID 同步单篇：已发布则写入，否则删除对应 beehive 文件。
func (s *SyncService) SyncSingle(ctx context.Context, articleID int64) error {
	var a models.Article
	if err := s.db.WithContext(ctx).Preload("Category").Preload("Tags").First(&a, articleID).Error; err != nil {
		return err
	}
	if a.Status != models.ArticleStatusPublished {
		return s.DeletePostFile(&a)
	}
	if _, err := s.writeArticleFile(ctx, &a); err != nil {
		return err
	}
	if s.rdb != nil {
		_ = s.rdb.Set(ctx, redisLastSyncKey, time.Now().UTC().Format(time.RFC3339Nano), 0).Err()
	}
	return nil
}

// DeletePostFile 删除该文章 ID 对应的全部 beehive-{id}-*.md（处理 slug 变更）。
func (s *SyncService) DeletePostFile(article *models.Article) error {
	if article == nil {
		return errors.New("article is nil")
	}
	return s.removePostsFilesForArticleID(article.ID)
}

func (s *SyncService) writeArticleFile(ctx context.Context, article *models.Article) (SyncAction, error) {
	_ = ctx
	cat := CategoryNameOrEmpty(article.Category)
	tags := SortedTagNames(article.Tags)
	payload, err := ArticleToHexoMarkdown(article, cat, tags)
	if err != nil {
		return "", err
	}
	name := GenerateHexoFileName(article)
	path := filepath.Join(s.postsDirAbs, name)

	hadFiles, err := s.hasBeehiveFilesForArticleID(article.ID)
	if err != nil {
		return "", err
	}
	if err := s.removePostsFilesForArticleID(article.ID); err != nil {
		return "", err
	}

	if err := atomicWriteFile(path, payload, 0o644); err != nil {
		return "", err
	}
	if hadFiles {
		return SyncActionUpdate, nil
	}
	return SyncActionCreate, nil
}

func (s *SyncService) hasBeehiveFilesForArticleID(id int64) (bool, error) {
	entries, err := os.ReadDir(s.postsDirAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "beehive-") || !strings.HasSuffix(name, ".md") {
			continue
		}
		parsedID, ok := parseBeehivePostFileID(name)
		if ok && parsedID == id {
			return true, nil
		}
	}
	return false, nil
}

func (s *SyncService) removePostsFilesForArticleID(id int64) error {
	entries, err := os.ReadDir(s.postsDirAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(s.postsDirAbs, 0o755)
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "beehive-") || !strings.HasSuffix(name, ".md") {
			continue
		}
		parsedID, ok := parseBeehivePostFileID(name)
		if !ok || parsedID != id {
			continue
		}
		_ = os.Remove(filepath.Join(s.postsDirAbs, name))
	}
	return nil
}

func parseBeehivePostFileID(filename string) (int64, bool) {
	// beehive-{id}-{slug}.md
	s := strings.TrimSuffix(filename, ".md")
	if !strings.HasPrefix(s, "beehive-") {
		return 0, false
	}
	rest := strings.TrimPrefix(s, "beehive-")
	idx := strings.Index(rest, "-")
	if idx <= 0 {
		return 0, false
	}
	idStr := rest[:idx]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (s *SyncService) cleanupOrphaned(ctx context.Context, active map[int64]struct{}) (int, error) {
	_ = ctx
	entries, err := os.ReadDir(s.postsDirAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, os.MkdirAll(s.postsDirAbs, 0o755)
		}
		return 0, err
	}
	removed := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "beehive-") || !strings.HasSuffix(name, ".md") {
			continue
		}
		id, ok := parseBeehivePostFileID(name)
		if !ok {
			continue
		}
		if _, keep := active[id]; keep {
			continue
		}
		if err := os.Remove(filepath.Join(s.postsDirAbs, name)); err != nil && !os.IsNotExist(err) {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".beehive-hexo-sync-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if perm != 0 {
		if err := os.Chmod(tmpName, perm); err != nil {
			return err
		}
	}
	return os.Rename(tmpName, path)
}

// RunHexoGenerate 在配置非空时执行静态站点生成命令。
func (s *SyncService) RunHexoGenerate(ctx context.Context) error {
	if len(s.generateArgs) == 0 {
		return errors.New("hexo generate_args is not configured")
	}
	cmd := exec.CommandContext(ctx, s.generateArgs[0], s.generateArgs[1:]...)
	cmd.Dir = s.generateWorkdir
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hexo generate: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// LocalBeehivePostCount 统计目录下 beehive-*.md 数量。
func (s *SyncService) LocalBeehivePostCount() (int, error) {
	entries, err := os.ReadDir(s.postsDirAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, "beehive-") && strings.HasSuffix(name, ".md") {
			n++
		}
	}
	return n, nil
}

// PublishedArticleCount 返回已发布文章数。
func (s *SyncService) PublishedArticleCount(ctx context.Context) (int64, error) {
	var c int64
	err := s.db.WithContext(ctx).Model(&models.Article{}).
		Where("status = ?", models.ArticleStatusPublished).
		Count(&c).Error
	return c, err
}

// LastSyncTime 读取上次成功全量/单篇同步写入 Redis 的时间。
func (s *SyncService) LastSyncTime(ctx context.Context) (time.Time, error) {
	if s.rdb == nil {
		return time.Time{}, nil
	}
	v, err := s.rdb.Get(ctx, redisLastSyncKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, v)
}

// PendingSync 判断是否存在更新时间晚于上次同步的已发布文章。
func (s *SyncService) PendingSync(ctx context.Context) (bool, error) {
	last, err := s.LastSyncTime(ctx)
	if err != nil {
		return false, err
	}
	if last.IsZero() {
		var c int64
		err = s.db.WithContext(ctx).Model(&models.Article{}).
			Where("status = ?", models.ArticleStatusPublished).
			Count(&c).Error
		if err != nil {
			return false, err
		}
		return c > 0, nil
	}
	var c int64
	err = s.db.WithContext(ctx).Model(&models.Article{}).
		Where("status = ? AND updated_at > ?", models.ArticleStatusPublished, last).
		Count(&c).Error
	if err != nil {
		return false, err
	}
	return c > 0, nil
}
