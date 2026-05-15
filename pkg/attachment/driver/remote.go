package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// RemoteDriver provides provider-neutral presigned URL generation for S3/OSS.
// RemoteDriver 为 S3/OSS 提供供应商无关的预签名 URL 生成。
type RemoteDriver struct {
	driverName      string
	bucket          string
	uploadBaseURL   string
	downloadBaseURL string
}

// NewRemoteDriver returns a DriverFactory that creates a RemoteDriver from a JSON config blob.
// Expected config: {"bucket": "...", "upload_base_url": "...", "download_base_url": "..."}.
// NewRemoteDriver 返回一个 DriverFactory，从 JSON 配置创建 RemoteDriver。
// 配置格式：{"bucket": "...", "upload_base_url": "...", "download_base_url": "..."}。
func NewRemoteDriver(driverName string) DriverFactory {
	return func(config json.RawMessage) (DriverBackend, error) {
		var cfg struct {
			Bucket          string `json:"bucket"`
			UploadBaseURL   string `json:"upload_base_url"`
			DownloadBaseURL string `json:"download_base_url"`
		}
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("%s driver config: %w", driverName, err)
		}
		return &RemoteDriver{
			driverName:      driverName,
			bucket:          strings.TrimSpace(cfg.Bucket),
			uploadBaseURL:   strings.TrimRight(strings.TrimSpace(cfg.UploadBaseURL), "/"),
			downloadBaseURL: strings.TrimRight(strings.TrimSpace(cfg.DownloadBaseURL), "/"),
		}, nil
	}
}

func (d *RemoteDriver) DriverName() string {
	return d.driverName
}

func (d *RemoteDriver) Save(context.Context, PutRequest) (StoredObject, error) {
	return StoredObject{}, ErrUnsupportedDriver
}

// PresignUpload generates a presigned PUT URL for direct client upload.
// PresignUpload 生成客户端直传的预签名 PUT URL。
func (d *RemoteDriver) PresignUpload(_ context.Context, req PresignRequest) (PresignResult, error) {
	if err := d.validate(); err != nil {
		return PresignResult{}, err
	}
	objectKey, err := cleanObjectKey(req.ObjectKey)
	if err != nil {
		return PresignResult{}, err
	}
	expiresAt := time.Now().Add(req.TTL)
	headers := map[string]string{
		"Content-Type": req.MimeType,
	}
	if req.Checksum != "" {
		headers["X-Content-Checksum"] = req.Checksum
	}
	return PresignResult{
		Bucket:    d.bucket,
		ObjectKey: objectKey,
		URL:       joinObjectURL(d.uploadBaseURL, objectKey, expiresAt),
		Method:    "PUT",
		Headers:   headers,
		ExpiresAt: expiresAt,
	}, nil
}

// PresignDownload generates a presigned GET URL with TTL.
// PresignDownload 生成带有 TTL 的预签名 GET URL。
func (d *RemoteDriver) PresignDownload(_ context.Context, objectKey string, ttl time.Duration) (PresignResult, error) {
	if err := d.validate(); err != nil {
		return PresignResult{}, err
	}
	clean, err := cleanObjectKey(objectKey)
	if err != nil {
		return PresignResult{}, err
	}
	expiresAt := time.Now().Add(ttl)
	return PresignResult{
		Bucket:    d.bucket,
		ObjectKey: clean,
		URL:       joinObjectURL(d.downloadBaseURL, clean, expiresAt),
		Method:    "GET",
		ExpiresAt: expiresAt,
	}, nil
}

func (d *RemoteDriver) LocalFilePath(string) (string, error) {
	return "", ErrUnsupportedDriver
}

// Delete is not supported for remote drivers in the current implementation.
// Delete 当前实现中 RemoteDriver 不支持服务端删除。
func (d *RemoteDriver) Delete(ctx context.Context, objectKey string) error {
	return ErrUnsupportedDriver
}

// HealthCheck verifies the remote config is present.
// HealthCheck 验证远端配置是否齐全。
func (d *RemoteDriver) HealthCheck(ctx context.Context) error {
	return d.validate()
}

func (d *RemoteDriver) validate() error {
	if d.bucket == "" || d.uploadBaseURL == "" || d.downloadBaseURL == "" {
		return fmt.Errorf("%s remote storage is not configured", d.driverName)
	}
	return nil
}
