package settings

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

// ErrInvalidSettings marks caller-supplied settings payload validation failures.
// ErrInvalidSettings 标记调用方提交的设置 payload 校验失败。
var ErrInvalidSettings = errors.New("invalid settings")

// Store loads and persists the singleton setting.application_settings row.
// Store 读写 setting.application_settings 单行。
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

// GetRevision returns revision only for cheap probes.
// GetRevision 仅返回 revision，用于轻量探测。
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

// EnsureSingleton inserts id=1 when no live row exists; unique_violation is treated as success.
// EnsureSingleton 在无活跃 id=1 行时插入；唯一约束冲突视为成功。
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
		if !hasBackfillValues(initial) {
			return nil
		}
		return s.backfillEmpty(ctx, initial)
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

func (s *Store) backfillEmpty(ctx context.Context, initial settingtypes.ApplicationSettings) error {
	if err := initial.Validate(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		next, changed := backfillEmptySettings(current, initial)
		if !changed {
			return nil
		}
		if err := next.Validate(); err != nil {
			return err
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
			return fmt.Errorf("backfill application settings: %w", err)
		}
		return nil
	})
}

// Save persists settings inside a row-lock transaction and bumps revision only when the normalized payload changes.
// Save 在行锁事务内持久化设置；仅当规范化后的 payload 发生变化时才递增 revision。
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
		current, err := settingtypes.ParsePayload(row.Payload)
		if err != nil {
			return err
		}
		rawCurrent, err := settingtypes.MarshalPayload(current)
		if err != nil {
			return err
		}
		if bytes.Equal(raw, rawCurrent) {
			savedRev = row.Revision
			return nil
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

// Patch locks the singleton row, merges the patch against the latest payload, and bumps revision only when the merged payload changes.
// Patch 锁定单行配置并合并补丁；仅当合并后的规范化 payload 发生变化时才递增 revision。
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
		rawCurrent, err := settingtypes.MarshalPayload(current)
		if err != nil {
			return err
		}
		if bytes.Equal(raw, rawCurrent) {
			saved = next
			savedRev = row.Revision
			return nil
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

func isPostgresUniqueViolation(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23505"
}

func hasBackfillValues(s settingtypes.ApplicationSettings) bool {
	s.Normalize()
	e := s.Email
	if e.Enabled ||
		strings.TrimSpace(e.Host) != "" ||
		e.Port != 587 ||
		strings.TrimSpace(e.Username) != "" ||
		strings.TrimSpace(e.Password) != "" ||
		strings.TrimSpace(e.From) != "" ||
		strings.TrimSpace(e.FromName) != "" ||
		strings.TrimSpace(e.TLS) != settingtypes.EmailTLSStartTLS {
		return true
	}
	g := s.GithubOAuth2
	return g.Enabled ||
		strings.TrimSpace(g.ClientID) != "" ||
		strings.TrimSpace(g.ClientSecret) != "" ||
		strings.TrimSpace(g.RedirectURL) != "" ||
		strings.TrimSpace(g.AuthURL) != settingtypes.DefaultGitHubAuthURL ||
		strings.TrimSpace(g.TokenURL) != settingtypes.DefaultGitHubTokenURL ||
		strings.TrimSpace(g.UserInfoURL) != settingtypes.DefaultGitHubUserInfoURL ||
		g.AllowNonGitHubEndpoints
}

func backfillEmptySettings(current, seed settingtypes.ApplicationSettings) (settingtypes.ApplicationSettings, bool) {
	current.Normalize()
	seed.Normalize()
	out := current
	changed := false

	emailEmpty := !current.Email.Enabled &&
		strings.TrimSpace(current.Email.Host) == "" &&
		strings.TrimSpace(current.Email.Username) == "" &&
		strings.TrimSpace(current.Email.Password) == "" &&
		strings.TrimSpace(current.Email.From) == "" &&
		strings.TrimSpace(current.Email.FromName) == ""
	if emailEmpty {
		if current.Email != seed.Email {
			out.Email = seed.Email
			changed = true
		}
	} else {
		changed = fillString(&out.Email.Host, seed.Email.Host) || changed
		changed = fillString(&out.Email.Username, seed.Email.Username) || changed
		changed = fillString(&out.Email.Password, seed.Email.Password) || changed
		changed = fillString(&out.Email.From, seed.Email.From) || changed
		changed = fillString(&out.Email.FromName, seed.Email.FromName) || changed
	}

	githubEmpty := !current.GithubOAuth2.Enabled &&
		strings.TrimSpace(current.GithubOAuth2.ClientID) == "" &&
		strings.TrimSpace(current.GithubOAuth2.ClientSecret) == "" &&
		strings.TrimSpace(current.GithubOAuth2.RedirectURL) == ""
	if githubEmpty {
		if current.GithubOAuth2 != seed.GithubOAuth2 {
			out.GithubOAuth2 = seed.GithubOAuth2
			changed = true
		}
	} else {
		changed = fillString(&out.GithubOAuth2.ClientID, seed.GithubOAuth2.ClientID) || changed
		changed = fillString(&out.GithubOAuth2.ClientSecret, seed.GithubOAuth2.ClientSecret) || changed
		changed = fillString(&out.GithubOAuth2.RedirectURL, seed.GithubOAuth2.RedirectURL) || changed
		changed = fillString(&out.GithubOAuth2.AuthURL, seed.GithubOAuth2.AuthURL) || changed
		changed = fillString(&out.GithubOAuth2.TokenURL, seed.GithubOAuth2.TokenURL) || changed
		changed = fillString(&out.GithubOAuth2.UserInfoURL, seed.GithubOAuth2.UserInfoURL) || changed
	}

	return out, changed
}

func fillString(dst *string, seed string) bool {
	if strings.TrimSpace(*dst) != "" || strings.TrimSpace(seed) == "" {
		return false
	}
	*dst = seed
	return true
}
