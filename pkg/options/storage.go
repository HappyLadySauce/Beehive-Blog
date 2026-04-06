package options

import (
	"github.com/spf13/pflag"
)

const (
	DefaultUploadDir   = "uploads"
	DefaultBaseURL     = "http://localhost:8081/uploads"
	DefaultMaxFileSize = int64(1000 * 1024 * 1024) // 1000MB
)

// StorageOptions 本地文件存储配置。
type StorageOptions struct {
	// UploadDir 上传文件根目录（相对或绝对路径），默认 "uploads"
	UploadDir string `json:"uploadDir" mapstructure:"uploadDir"`
	// BaseURL 文件访问 URL 前缀（含 scheme/host），默认 "http://localhost:8081/uploads"
	BaseURL string `json:"baseUrl" mapstructure:"baseUrl"`
	// MaxFileSize 单文件最大字节数，默认 1000MB
	MaxFileSize int64 `json:"maxFileSize" mapstructure:"maxFileSize"`
}

// NewStorageOptions returns StorageOptions with sensible defaults.
func NewStorageOptions() *StorageOptions {
	return &StorageOptions{
		UploadDir:   DefaultUploadDir,
		BaseURL:     DefaultBaseURL,
		MaxFileSize: DefaultMaxFileSize,
	}
}

// AddFlags registers storage-related flags.
func (s *StorageOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.UploadDir, "uploadDir", s.UploadDir, "Directory for uploaded files")
	fs.StringVar(&s.BaseURL, "storageBaseUrl", s.BaseURL, "Base URL prefix for accessing uploaded files")
	fs.Int64Var(&s.MaxFileSize, "maxFileSize", s.MaxFileSize, "Maximum upload file size in bytes")
}
