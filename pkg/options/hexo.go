package options

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

// HexoOptions 控制数据库文章同步到 Hexo source/_posts 的行为。
type HexoOptions struct {
	PostsDir string `mapstructure:"posts_dir"`
	// AutoSync 为 true 时可在文章 CRUD 后触发单篇同步（当前版本默认关闭，由配置控制）。
	AutoSync bool `mapstructure:"auto_sync"`
	// GenerateWorkdir 执行生成命令时的工作目录，默认 ui/hexo。
	GenerateWorkdir string `mapstructure:"generate_workdir"`
	// CleanArgs 非空时在 RunHexoRebuild 中先于 generate 执行，例如 ["hexo","clean"]（需 PATH 中有 hexo）。
	CleanArgs []string `mapstructure:"clean_args"`
	// GenerateArgs 非空时在 RunHexoRebuild 中执行生成步骤，例如 ["hexo","generate"]。
	GenerateArgs []string `mapstructure:"generate_args"`
	// RebuildAfterAutoSync 为 true 时，单篇自动同步成功后执行 RunHexoClean+Generate（成本高，默认 false）。
	RebuildAfterAutoSync bool `mapstructure:"rebuild_after_auto_sync"`
}

const (
	DefaultHexoPostsDir        = "ui/hexo/source/_posts"
	DefaultHexoGenerateWorkdir = "ui/hexo"
)

// NewHexoOptions 返回带默认值的 Hexo 配置（手动同步友好：AutoSync 默认 false）。
func NewHexoOptions() *HexoOptions {
	return &HexoOptions{
		PostsDir:             DefaultHexoPostsDir,
		AutoSync:             false,
		GenerateWorkdir:      DefaultHexoGenerateWorkdir,
		CleanArgs:            nil,
		GenerateArgs:         nil,
		RebuildAfterAutoSync: false,
	}
}

// Validate 校验并补全 Hexo 路径默认值。
func (h *HexoOptions) Validate() error {
	if h == nil {
		return nil
	}
	if strings.TrimSpace(h.PostsDir) == "" {
		h.PostsDir = DefaultHexoPostsDir
	}
	h.PostsDir = filepath.Clean(h.PostsDir)
	if h.PostsDir == "." {
		return errors.New("hexo.posts_dir is invalid")
	}
	if strings.TrimSpace(h.GenerateWorkdir) == "" {
		h.GenerateWorkdir = DefaultHexoGenerateWorkdir
	}
	h.GenerateWorkdir = filepath.Clean(h.GenerateWorkdir)
	if h.GenerateWorkdir == "." {
		return errors.New("hexo.generate_workdir is invalid")
	}
	return nil
}

func (h *HexoOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&h.PostsDir, "postsDir", "", h.PostsDir, "Hexo posts directory")
	fs.BoolVarP(&h.AutoSync, "autoSync", "", h.AutoSync, "Auto sync posts to Hexo")
	fs.StringVarP(&h.GenerateWorkdir, "generateWorkdir", "", h.GenerateWorkdir, "Hexo generate work directory")
	fs.StringSliceVarP(&h.CleanArgs, "cleanArgs", "", h.CleanArgs, "Hexo clean argv prefix e.g. hexo clean")
	fs.StringSliceVarP(&h.GenerateArgs, "generateArgs", "", h.GenerateArgs, "Hexo generate argv prefix e.g. hexo generate")
	fs.BoolVarP(&h.RebuildAfterAutoSync, "rebuildAfterAutoSync", "", h.RebuildAfterAutoSync, "Run hexo clean+generate after each auto SyncSingle")
}
