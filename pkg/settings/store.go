package settings

import (
	"context"
	"fmt"
	"time"

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
