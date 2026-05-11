package options

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

const (
	AttachmentStorageLocal = "local"
	AttachmentStorageS3    = "s3"
	AttachmentStorageOSS   = "oss"
)

// AttachmentRemoteOptions holds provider-neutral knobs for remote direct-upload flows.
// AttachmentRemoteOptions 保存远端直传流程的通用配置。
type AttachmentRemoteOptions struct {
	Bucket          string `json:"bucket"            mapstructure:"bucket"`
	UploadBaseURL   string `json:"upload-base-url"   mapstructure:"upload-base-url"`
	DownloadBaseURL string `json:"download-base-url" mapstructure:"download-base-url"`
}

// AttachmentOptions holds attachment storage, validation and signing defaults.
// AttachmentOptions 保存附件存储、校验与签名相关默认配置。
type AttachmentOptions struct {
	DefaultStorage      string                  `json:"default-storage"       mapstructure:"default-storage"`
	LocalRoot           string                  `json:"local-root"            mapstructure:"local-root"`
	MaxBytes            int64                   `json:"max-bytes"             mapstructure:"max-bytes"`
	AllowedMIMEPrefixes []string                `json:"allowed-mime-prefixes" mapstructure:"allowed-mime-prefixes"`
	PresignTTL          time.Duration           `json:"presign-ttl"           mapstructure:"presign-ttl"`
	S3                  AttachmentRemoteOptions `json:"s3"                    mapstructure:"s3"`
	OSS                 AttachmentRemoteOptions `json:"oss"                   mapstructure:"oss"`
}

// NewAttachmentOptions returns safe defaults for local development.
// NewAttachmentOptions 返回适合本地开发的安全默认值。
func NewAttachmentOptions() *AttachmentOptions {
	return &AttachmentOptions{
		DefaultStorage: AttachmentStorageLocal,
		LocalRoot:      "data/attachments",
		MaxBytes:       10 << 20,
		AllowedMIMEPrefixes: []string{
			"image/",
			"application/pdf",
		},
		PresignTTL: 15 * time.Minute,
	}
}

// Validate checks attachment configuration consistency.
// Validate 校验附件配置的一致性。
func (o *AttachmentOptions) Validate() error {
	if o == nil {
		return fmt.Errorf("attachment options is nil")
	}
	var err error
	if !attachmentStorageKnown(o.DefaultStorage) {
		err = errors.Join(err, fmt.Errorf("attachment default-storage must be one of local, s3, oss, got %q", o.DefaultStorage))
	}
	if strings.TrimSpace(o.LocalRoot) == "" {
		err = errors.Join(err, fmt.Errorf("attachment local-root is required"))
	}
	if o.MaxBytes <= 0 {
		err = errors.Join(err, fmt.Errorf("attachment max-bytes must be > 0, got %d", o.MaxBytes))
	}
	if len(o.AllowedMIMEPrefixes) == 0 {
		err = errors.Join(err, fmt.Errorf("attachment allowed-mime-prefixes must not be empty"))
	}
	for _, prefix := range o.AllowedMIMEPrefixes {
		if strings.TrimSpace(prefix) == "" {
			err = errors.Join(err, fmt.Errorf("attachment allowed-mime-prefixes must not contain empty values"))
		}
	}
	if o.PresignTTL <= 0 {
		err = errors.Join(err, fmt.Errorf("attachment presign-ttl must be > 0, got %s", o.PresignTTL))
	}
	err = errors.Join(err, validateRemoteOptions("s3", o.S3))
	err = errors.Join(err, validateRemoteOptions("oss", o.OSS))
	return err
}

// AddFlags registers attachment flags on the supplied FlagSet.
// AddFlags 将附件相关命令行标志注册到给定的 FlagSet。
func (o *AttachmentOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.DefaultStorage, "attachment-default-storage", o.DefaultStorage, "Attachment storage backend: local, s3, or oss")
	fs.StringVar(&o.LocalRoot, "attachment-local-root", o.LocalRoot, "Root directory for local attachment files")
	fs.Int64Var(&o.MaxBytes, "attachment-max-bytes", o.MaxBytes, "Maximum upload size in bytes")
	fs.StringSliceVar(&o.AllowedMIMEPrefixes, "attachment-allowed-mime-prefixes", o.AllowedMIMEPrefixes, "Allowed MIME prefixes, e.g. image/,application/pdf")
	fs.DurationVar(&o.PresignTTL, "attachment-presign-ttl", o.PresignTTL, "Remote upload/download pre-sign lifetime")
	fs.StringVar(&o.S3.Bucket, "attachment-s3-bucket", o.S3.Bucket, "S3 bucket for remote attachments")
	fs.StringVar(&o.S3.UploadBaseURL, "attachment-s3-upload-base-url", o.S3.UploadBaseURL, "S3-compatible direct upload base URL")
	fs.StringVar(&o.S3.DownloadBaseURL, "attachment-s3-download-base-url", o.S3.DownloadBaseURL, "S3-compatible public/download base URL")
	fs.StringVar(&o.OSS.Bucket, "attachment-oss-bucket", o.OSS.Bucket, "OSS bucket for remote attachments")
	fs.StringVar(&o.OSS.UploadBaseURL, "attachment-oss-upload-base-url", o.OSS.UploadBaseURL, "OSS-compatible direct upload base URL")
	fs.StringVar(&o.OSS.DownloadBaseURL, "attachment-oss-download-base-url", o.OSS.DownloadBaseURL, "OSS-compatible public/download base URL")
}

func attachmentStorageKnown(v string) bool {
	switch v {
	case AttachmentStorageLocal, AttachmentStorageS3, AttachmentStorageOSS:
		return true
	default:
		return false
	}
}

func validateRemoteOptions(name string, o AttachmentRemoteOptions) error {
	if strings.TrimSpace(o.Bucket) == "" && strings.TrimSpace(o.UploadBaseURL) == "" && strings.TrimSpace(o.DownloadBaseURL) == "" {
		return nil
	}
	var err error
	if strings.TrimSpace(o.Bucket) == "" {
		err = errors.Join(err, fmt.Errorf("attachment %s.bucket is required when %s remote options are set", name, name))
	}
	if strings.TrimSpace(o.UploadBaseURL) == "" {
		err = errors.Join(err, fmt.Errorf("attachment %s.upload-base-url is required when %s remote options are set", name, name))
	}
	if strings.TrimSpace(o.DownloadBaseURL) == "" {
		err = errors.Join(err, fmt.Errorf("attachment %s.download-base-url is required when %s remote options are set", name, name))
	}
	return err
}
