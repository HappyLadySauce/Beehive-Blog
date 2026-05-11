package settings

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

// Store loads and persists the singleton setting.application_settings row.
// Store 负责读写 setting.application_settings 单行。
type Store struct {
	db *gorm.DB
}

// ErrInvalidSettings marks caller-supplied settings payload validation failures.
// ErrInvalidSettings 标记调用方提交的设置 payload 校验失败。
var ErrInvalidSettings = errors.New("invalid settings")

// NewStore builds a Store backed by the given GORM handle.
// NewStore 使用给定 GORM 句柄构造 Store。
func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

// Load returns decoded settings and the current revision for id=1.
// Load 返回 id=1 的解码设置与当前 revision。
func (s *Store) Load(ctx context.Context) (settingtypes.ApplicationSettings, int64, error) {
	var row model.ApplicationSetting
	if err := s.db.WithContext(ctx).Where("id = ?", 1).First(&row).Error; err != nil {
		return settingtypes.ApplicationSettings{}, 0, fmt.Errorf("load application settings: %w", err)
	}
	out, err := settingtypes.ParsePayload(row.Payload)
	if err != nil {
		return settingtypes.ApplicationSettings{}, 0, err
	}
	return out, row.Revision, nil
}

// GetRevision returns revision only (cheap probe for hot reload).
// GetRevision 仅返回 revision（热加载轻量探测）。
func (s *Store) GetRevision(ctx context.Context) (int64, error) {
	var row model.ApplicationSetting
	if err := s.db.WithContext(ctx).
		Model(&model.ApplicationSetting{}).
		Select("revision").
		Where("id = ?", 1).
		Take(&row).Error; err != nil {
		return 0, fmt.Errorf("read settings revision: %w", err)
	}
	return row.Revision, nil
}

// EnsureSingleton inserts id=1 when no live row exists; unique_violation is treated as success (concurrent seed).
// EnsureSingleton 在无活跃 id=1 行时插入；唯一约束冲突视为成功（并发补种）。
func (s *Store) EnsureSingleton(ctx context.Context, initial settingtypes.ApplicationSettings) error {
	if err := initial.Validate(); err != nil {
		return err
	}
	raw, err := settingtypes.MarshalPayload(initial)
	if err != nil {
		return err
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&model.ApplicationSetting{}).Where("id = ?", 1).Count(&count).Error; err != nil {
		return fmt.Errorf("ensure application settings: count id=1: %w", err)
	}
	if count > 0 {
		return nil
	}

	now := time.Now()
	row := model.ApplicationSetting{
		ID:        1,
		Revision:  1,
		Payload:   raw,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		if isPostgresUniqueViolation(err) {
			return nil
		}
		return fmt.Errorf("ensure application settings: insert singleton: %w", err)
	}
	return nil
}

func isPostgresUniqueViolation(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23505"
}

// Save persists merged settings and increments revision inside a row lock transaction.
// Save 在行锁事务内持久化合并后的设置并递增 revision。
func (s *Store) Save(ctx context.Context, next settingtypes.ApplicationSettings) (int64, error) {
	if err := next.Validate(); err != nil {
		return 0, err
	}
	raw, err := settingtypes.MarshalPayload(next)
	if err != nil {
		return 0, err
	}
	var savedRev int64
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row model.ApplicationSetting
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", 1).
			First(&row).Error; err != nil {
			return fmt.Errorf("lock application settings: %w", err)
		}
		nextRev := row.Revision + 1
		now := time.Now()
		if err := tx.Model(&row).Updates(map[string]any{
			"payload":    raw,
			"revision":   nextRev,
			"updated_at": now,
		}).Error; err != nil {
			return fmt.Errorf("save application settings: %w", err)
		}
		savedRev = nextRev
		return nil
	}); err != nil {
		return 0, err
	}
	return savedRev, nil
}

// Patch locks the singleton row, merges the patch against the latest payload, and increments revision.
// Patch 锁定单行配置，基于最新 payload 合并补丁，并递增 revision。
func (s *Store) Patch(ctx context.Context, patch *settingtypes.SettingsPatchRequest) (settingtypes.ApplicationSettings, int64, error) {
	var saved settingtypes.ApplicationSettings
	var savedRev int64
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row model.ApplicationSetting
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", 1).
			First(&row).Error; err != nil {
			return fmt.Errorf("lock application settings: %w", err)
		}
		current, err := settingtypes.ParsePayload(row.Payload)
		if err != nil {
			return err
		}
		next, err := settingtypes.MergePatch(current, patch)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidSettings, err)
		}
		raw, err := settingtypes.MarshalPayload(next)
		if err != nil {
			return err
		}
		nextRev := row.Revision + 1
		now := time.Now()
		if err := tx.Model(&row).Updates(map[string]any{
			"payload":    raw,
			"revision":   nextRev,
			"updated_at": now,
		}).Error; err != nil {
			return fmt.Errorf("patch application settings: %w", err)
		}
		saved = next
		savedRev = nextRev
		return nil
	}); err != nil {
		return settingtypes.ApplicationSettings{}, 0, err
	}
	return saved, savedRev, nil
}
