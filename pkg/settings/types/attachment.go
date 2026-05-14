package types

import (
	"errors"
	"fmt"
	"strings"
)

// Attachment storage backend constants.
// 附件存储后端常量。
const (
	AttachmentStorageLocal = "local"
	AttachmentStorageS3    = "s3"
	AttachmentStorageOSS   = "oss"

	defaultAttachmentMaxBytes   = 10 << 20
	defaultAttachmentPresignTTL = 900 // seconds (15 min)
	defaultAttachmentLocalRoot  = "data/attachments"
)

// AttachmentRemoteSettings holds remote direct-upload configuration persisted in the database.
// AttachmentRemoteSettings 保存于数据库的远端直传配置。
type AttachmentRemoteSettings struct {
	Bucket          string `json:"bucket"`
	UploadBaseURL   string `json:"upload_base_url"`
	DownloadBaseURL string `json:"download_base_url"`
}

// AttachmentSettings holds attachment storage and validation defaults persisted in the database.
// AttachmentSettings 保存于数据库的附件存储与校验默认配置。
type AttachmentSettings struct {
	DefaultStorage      string                  `json:"default_storage"`
	LocalRoot           string                  `json:"local_root"`
	MaxBytes            int64                   `json:"max_bytes"`
	AllowedMIMEPrefixes []string                `json:"allowed_mime_prefixes"`
	PresignTTLSeconds   int64                   `json:"presign_ttl_seconds"`
	S3                  AttachmentRemoteSettings `json:"s3"`
	OSS                 AttachmentRemoteSettings `json:"oss"`
}

// AttachmentRemotePatch uses pointers so omitted JSON fields leave existing values unchanged.
// AttachmentRemotePatch 使用指针，JSON 省略的字段保留原值。
type AttachmentRemotePatch struct {
	Bucket          *string `json:"bucket"`
	UploadBaseURL   *string `json:"upload_base_url"`
	DownloadBaseURL *string `json:"download_base_url"`
}

// AttachmentPatch uses pointers so omitted JSON fields leave existing values unchanged.
// AttachmentPatch 使用指针，JSON 省略的字段保留原值。
type AttachmentPatch struct {
	DefaultStorage      *string                `json:"default_storage"`
	LocalRoot           *string                `json:"local_root"`
	MaxBytes            *int64                 `json:"max_bytes"`
	AllowedMIMEPrefixes *[]string              `json:"allowed_mime_prefixes"`
	PresignTTLSeconds   *int64                 `json:"presign_ttl_seconds"`
	S3                  *AttachmentRemotePatch `json:"s3"`
	OSS                 *AttachmentRemotePatch `json:"oss"`
}

// Normalize fills defaults for omitted fields after decode.
// Normalize 在解码后为缺省字段填充默认值。
func (s *AttachmentSettings) Normalize() {
	if !attachmentStorageKnown(s.DefaultStorage) {
		s.DefaultStorage = AttachmentStorageLocal
	}
	if strings.TrimSpace(s.LocalRoot) == "" {
		s.LocalRoot = defaultAttachmentLocalRoot
	}
	if s.MaxBytes <= 0 {
		s.MaxBytes = defaultAttachmentMaxBytes
	}
	if s.PresignTTLSeconds <= 0 {
		s.PresignTTLSeconds = defaultAttachmentPresignTTL
	}
	if len(s.AllowedMIMEPrefixes) == 0 {
		s.AllowedMIMEPrefixes = []string{"image/", "application/pdf"}
	}
}

func validateAttachments(s *AttachmentSettings) error {
	var err error
	if !attachmentStorageKnown(s.DefaultStorage) {
		err = errors.Join(err, fmt.Errorf("attachment.default_storage must be one of local, s3, oss, got %q", s.DefaultStorage))
	}
	if strings.TrimSpace(s.LocalRoot) == "" {
		err = errors.Join(err, fmt.Errorf("attachment.local_root is required"))
	}
	if s.MaxBytes <= 0 {
		err = errors.Join(err, fmt.Errorf("attachment.max_bytes must be > 0, got %d", s.MaxBytes))
	}
	if len(s.AllowedMIMEPrefixes) == 0 {
		err = errors.Join(err, fmt.Errorf("attachment.allowed_mime_prefixes must not be empty"))
	}
	for _, prefix := range s.AllowedMIMEPrefixes {
		if strings.TrimSpace(prefix) == "" {
			err = errors.Join(err, fmt.Errorf("attachment.allowed_mime_prefixes must not contain empty values"))
		}
	}
	if s.PresignTTLSeconds <= 0 {
		err = errors.Join(err, fmt.Errorf("attachment.presign_ttl_seconds must be > 0, got %d", s.PresignTTLSeconds))
	}
	err = errors.Join(err, validateAttachmentRemote("s3", s.S3))
	err = errors.Join(err, validateAttachmentRemote("oss", s.OSS))
	return err
}

func validateAttachmentRemote(name string, o AttachmentRemoteSettings) error {
	if strings.TrimSpace(o.Bucket) == "" && strings.TrimSpace(o.UploadBaseURL) == "" && strings.TrimSpace(o.DownloadBaseURL) == "" {
		return nil
	}
	var err error
	if strings.TrimSpace(o.Bucket) == "" {
		err = errors.Join(err, fmt.Errorf("attachment.%s.bucket is required when %s remote options are set", name, name))
	}
	if strings.TrimSpace(o.UploadBaseURL) == "" {
		err = errors.Join(err, fmt.Errorf("attachment.%s.upload_base_url is required when %s remote options are set", name, name))
	}
	if strings.TrimSpace(o.DownloadBaseURL) == "" {
		err = errors.Join(err, fmt.Errorf("attachment.%s.download_base_url is required when %s remote options are set", name, name))
	}
	return err
}

func attachmentStorageKnown(v string) bool {
	switch v {
	case AttachmentStorageLocal, AttachmentStorageS3, AttachmentStorageOSS:
		return true
	default:
		return false
	}
}
