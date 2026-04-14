// Package hexocfg merges YAML hexo_dir with DB settings (group=hexo) for sync and rebuild.
package hexocfg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/v1/pkg/options"
	"gorm.io/gorm"
)

// Settings keys (settings.key is globally unique).
const (
	KeyHexoDir              = "hexo.hexo_dir" // read-only from YAML; not stored in DB
	KeyAutoSync             = "hexo.auto_sync"
	KeyCleanArgs            = "hexo.clean_args"
	KeyGenerateArgs         = "hexo.generate_args"
	KeyRebuildAfterAutoSync = "hexo.rebuild_after_auto_sync"
)

// EffectiveHexo is the resolved config for Hexo sync and subprocess invocations.
type EffectiveHexo struct {
	PostsDirAbs          string
	GenerateWorkdirAbs   string
	AutoSync             bool
	CleanArgs            []string
	GenerateArgs         []string
	RebuildAfterAutoSync bool
}

// LoadEffective reads group=hexo from DB and merges with file-based HexoDir (absolute paths).
func LoadEffective(ctx context.Context, db *gorm.DB, file *options.HexoOptions) (*EffectiveHexo, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	if file == nil {
		return nil, errors.New("hexo file options is nil")
	}
	hexoDir := strings.TrimSpace(file.HexoDir)
	if hexoDir == "" {
		hexoDir = options.DefaultHexoDir
	}
	hexoDir = filepath.Clean(hexoDir)
	genAbs, err := filepath.Abs(hexoDir)
	if err != nil {
		return nil, fmt.Errorf("resolve hexo_dir: %w", err)
	}
	postsAbs := filepath.Join(genAbs, "source", "_posts")

	var rows []models.Setting
	if err := db.WithContext(ctx).Where(`"group" = ?`, models.SettingGroupHexo).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("load hexo settings: %w", err)
	}
	kv := make(map[string]string, len(rows))
	for _, r := range rows {
		kv[r.Key] = r.Value
	}

	out := &EffectiveHexo{
		PostsDirAbs:          postsAbs,
		GenerateWorkdirAbs:   genAbs,
		AutoSync:             parseBool(kv[KeyAutoSync], false),
		CleanArgs:            nil,
		GenerateArgs:         nil,
		RebuildAfterAutoSync: parseBool(kv[KeyRebuildAfterAutoSync], false),
	}
	var e error
	out.CleanArgs, e = parseStringSliceArg(kv[KeyCleanArgs])
	if e != nil {
		return nil, fmt.Errorf("hexo.clean_args: %w", e)
	}
	out.GenerateArgs, e = parseStringSliceArg(kv[KeyGenerateArgs])
	if e != nil {
		return nil, fmt.Errorf("hexo.generate_args: %w", e)
	}
	return out, nil
}

func parseBool(s string, def bool) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return def
	}
	return s == "true" || s == "1" || s == "yes"
}

// parseStringSliceArg returns nil for empty / "[]" / unset; otherwise JSON array of strings.
func parseStringSliceArg(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil, nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, nil
	}
	for i, s := range arr {
		if strings.TrimSpace(s) == "" {
			return nil, fmt.Errorf("element %d is empty", i)
		}
	}
	if strings.TrimSpace(arr[0]) == "" {
		return nil, errors.New("argv[0] is empty")
	}
	return arr, nil
}

// ValidateHexoSettingValue validates a single key/value for PUT /settings/hexo.
func ValidateHexoSettingValue(key, value string) error {
	switch key {
	case KeyHexoDir:
		return errors.New("hexo.hexo_dir is read-only")
	case KeyAutoSync, KeyRebuildAfterAutoSync:
		v := strings.TrimSpace(strings.ToLower(value))
		if v != "true" && v != "false" {
			return fmt.Errorf("invalid boolean for %s", key)
		}
		return nil
	case KeyCleanArgs, KeyGenerateArgs:
		v := strings.TrimSpace(value)
		if v == "" || v == "[]" {
			return nil
		}
		_, err := parseStringSliceArg(v)
		return err
	default:
		return fmt.Errorf("unknown hexo setting key %q", key)
	}
}

// HexoDirDisplay returns the configured hexo root for API read-only merge (relative as in YAML).
func HexoDirDisplay(file *options.HexoOptions) string {
	if file == nil {
		return ""
	}
	d := strings.TrimSpace(file.HexoDir)
	if d == "" {
		d = options.DefaultHexoDir
	}
	return filepath.Clean(d)
}
