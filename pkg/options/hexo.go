package options

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

// HexoOptions holds only deployment-level Hexo root from YAML; behavior flags live in DB (group=hexo).
type HexoOptions struct {
	// HexoDir is the Hexo project root (where package.json / _config.yml live). Default ui/hexo.
	HexoDir string `mapstructure:"hexo_dir"`
}

// DefaultHexoDir is the default hexo root relative to the process working directory.
const DefaultHexoDir = "ui/hexo"

// NewHexoOptions returns defaults for file-based Hexo options.
func NewHexoOptions() *HexoOptions {
	return &HexoOptions{
		HexoDir: DefaultHexoDir,
	}
}

// Validate normalizes HexoDir (non-empty, clean path).
func (h *HexoOptions) Validate() error {
	if h == nil {
		return nil
	}
	if strings.TrimSpace(h.HexoDir) == "" {
		h.HexoDir = DefaultHexoDir
	}
	h.HexoDir = filepath.Clean(h.HexoDir)
	if h.HexoDir == "." {
		return errors.New("hexo.hexo_dir is invalid")
	}
	return nil
}

func (h *HexoOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&h.HexoDir, "hexoDir", "", h.HexoDir, "Hexo project root directory (e.g. ui/hexo)")
}
