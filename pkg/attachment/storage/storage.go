package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

var (
	ErrUnsupportedBackend = errors.New("attachment storage backend is unsupported")
	ErrUnsafeObjectKey    = errors.New("attachment object key is unsafe")
)

// PutRequest describes a server-side upload write.
// PutRequest 描述服务端上传写入请求。
type PutRequest struct {
	ObjectKey string
	Reader    io.Reader
	Size      int64
}

// StoredObject describes a stored object after write.
// StoredObject 描述写入后的存储对象。
type StoredObject struct {
	LocalPath string
	ETag      string
	Checksum  string
}

// PresignRequest describes a direct-upload or direct-download signing request.
// PresignRequest 描述直传或直读签名请求。
type PresignRequest struct {
	ObjectKey string
	MimeType  string
	Checksum  string
	Size      int64
	TTL       time.Duration
}

// PresignResult returns provider-neutral upload/download instructions.
// PresignResult 返回供应商无关的上传/下载指令。
type PresignResult struct {
	Bucket    string
	ObjectKey string
	URL       string
	Method    string
	Headers   map[string]string
	ExpiresAt time.Time
}

// Backend is the storage boundary used by the attachment service.
// Backend 是附件服务使用的存储边界。
type Backend interface {
	StorageType() string
	Save(ctx context.Context, req PutRequest) (StoredObject, error)
	PresignUpload(ctx context.Context, req PresignRequest) (PresignResult, error)
	PresignDownload(ctx context.Context, objectKey string, ttl time.Duration) (PresignResult, error)
	LocalFilePath(localPath string) (string, error)
}

// Registry resolves storage backends by storage_type.
// Registry 按 storage_type 解析存储后端。
type Registry struct {
	backends map[string]Backend
}

// NewRegistry builds a storage registry from attachment options.
// NewRegistry 基于附件配置构造存储后端注册表。
func NewRegistry(opts *options.AttachmentOptions) (*Registry, error) {
	if opts == nil {
		return nil, fmt.Errorf("attachment options is nil")
	}
	local, err := NewLocalBackend(opts.LocalRoot)
	if err != nil {
		return nil, err
	}
	backends := map[string]Backend{
		options.AttachmentStorageLocal: local,
		options.AttachmentStorageS3:    NewRemoteBackend(options.AttachmentStorageS3, opts.S3),
		options.AttachmentStorageOSS:   NewRemoteBackend(options.AttachmentStorageOSS, opts.OSS),
	}
	return &Registry{backends: backends}, nil
}

// Backend returns the configured backend for storageType.
// Backend 返回 storageType 对应的存储后端。
func (r *Registry) Backend(storageType string) (Backend, error) {
	if r == nil {
		return nil, fmt.Errorf("storage registry is nil")
	}
	backend, ok := r.backends[storageType]
	if !ok || backend == nil {
		return nil, ErrUnsupportedBackend
	}
	return backend, nil
}

// LocalBackend stores objects under a configured local root.
// LocalBackend 将对象存储在配置的本地根目录下。
type LocalBackend struct {
	root string
}

// NewLocalBackend creates a local storage backend.
// NewLocalBackend 创建本地存储后端。
func NewLocalBackend(root string) (*LocalBackend, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, fmt.Errorf("local attachment root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve local attachment root: %w", err)
	}
	return &LocalBackend{root: abs}, nil
}

func (b *LocalBackend) StorageType() string {
	return options.AttachmentStorageLocal
}

func (b *LocalBackend) Save(ctx context.Context, req PutRequest) (StoredObject, error) {
	if req.Reader == nil {
		return StoredObject{}, fmt.Errorf("attachment reader is nil")
	}
	localPath, err := cleanObjectKey(req.ObjectKey)
	if err != nil {
		return StoredObject{}, err
	}
	abs, err := b.LocalFilePath(localPath)
	if err != nil {
		return StoredObject{}, err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return StoredObject{}, fmt.Errorf("create attachment directory: %w", err)
	}
	out, err := os.OpenFile(abs, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return StoredObject{}, fmt.Errorf("create attachment file: %w", err)
	}
	defer out.Close()

	hasher := sha256.New()
	written, err := io.Copy(out, io.TeeReader(req.Reader, hasher))
	if err != nil {
		_ = os.Remove(abs)
		return StoredObject{}, fmt.Errorf("write attachment file: %w", err)
	}
	if req.Size >= 0 && written != req.Size {
		_ = os.Remove(abs)
		return StoredObject{}, fmt.Errorf("attachment size mismatch: expected %d bytes, wrote %d", req.Size, written)
	}
	sum := hex.EncodeToString(hasher.Sum(nil))
	return StoredObject{
		LocalPath: localPath,
		ETag:      sum,
		Checksum:  "sha256:" + sum,
	}, nil
}

func (b *LocalBackend) PresignUpload(context.Context, PresignRequest) (PresignResult, error) {
	return PresignResult{}, ErrUnsupportedBackend
}

func (b *LocalBackend) PresignDownload(context.Context, string, time.Duration) (PresignResult, error) {
	return PresignResult{}, ErrUnsupportedBackend
}

func (b *LocalBackend) LocalFilePath(localPath string) (string, error) {
	clean, err := cleanObjectKey(localPath)
	if err != nil {
		return "", err
	}
	abs := filepath.Join(b.root, filepath.FromSlash(clean))
	rel, err := filepath.Rel(b.root, abs)
	if err != nil {
		return "", fmt.Errorf("resolve attachment path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", ErrUnsafeObjectKey
	}
	return abs, nil
}

// RemoteBackend provides provider-neutral direct URL boundaries for s3/oss.
// RemoteBackend 为 s3/oss 提供供应商无关的直传 URL 边界。
type RemoteBackend struct {
	storageType     string
	bucket          string
	uploadBaseURL   string
	downloadBaseURL string
}

// NewRemoteBackend creates a provider-neutral remote backend.
// NewRemoteBackend 创建供应商无关的远端后端。
func NewRemoteBackend(storageType string, opts options.AttachmentRemoteOptions) *RemoteBackend {
	return &RemoteBackend{
		storageType:     storageType,
		bucket:          strings.TrimSpace(opts.Bucket),
		uploadBaseURL:   strings.TrimRight(strings.TrimSpace(opts.UploadBaseURL), "/"),
		downloadBaseURL: strings.TrimRight(strings.TrimSpace(opts.DownloadBaseURL), "/"),
	}
}

func (b *RemoteBackend) StorageType() string {
	return b.storageType
}

func (b *RemoteBackend) Save(context.Context, PutRequest) (StoredObject, error) {
	return StoredObject{}, ErrUnsupportedBackend
}

func (b *RemoteBackend) PresignUpload(_ context.Context, req PresignRequest) (PresignResult, error) {
	if err := b.validate(); err != nil {
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
		Bucket:    b.bucket,
		ObjectKey: objectKey,
		URL:       joinObjectURL(b.uploadBaseURL, objectKey, expiresAt),
		Method:    "PUT",
		Headers:   headers,
		ExpiresAt: expiresAt,
	}, nil
}

func (b *RemoteBackend) PresignDownload(_ context.Context, objectKey string, ttl time.Duration) (PresignResult, error) {
	if err := b.validate(); err != nil {
		return PresignResult{}, err
	}
	clean, err := cleanObjectKey(objectKey)
	if err != nil {
		return PresignResult{}, err
	}
	expiresAt := time.Now().Add(ttl)
	return PresignResult{
		Bucket:    b.bucket,
		ObjectKey: clean,
		URL:       joinObjectURL(b.downloadBaseURL, clean, expiresAt),
		Method:    "GET",
		ExpiresAt: expiresAt,
	}, nil
}

func (b *RemoteBackend) LocalFilePath(string) (string, error) {
	return "", ErrUnsupportedBackend
}

func (b *RemoteBackend) validate() error {
	if b.bucket == "" || b.uploadBaseURL == "" || b.downloadBaseURL == "" {
		return fmt.Errorf("%s attachment remote storage is not configured", b.storageType)
	}
	return nil
}

func cleanObjectKey(objectKey string) (string, error) {
	objectKey = strings.TrimSpace(strings.ReplaceAll(objectKey, "\\", "/"))
	if strings.HasPrefix(objectKey, "/") {
		return "", ErrUnsafeObjectKey
	}
	for _, part := range strings.Split(objectKey, "/") {
		if part == ".." {
			return "", ErrUnsafeObjectKey
		}
	}
	objectKey = strings.TrimPrefix(path.Clean("/"+objectKey), "/")
	if objectKey == "." || objectKey == "" || strings.HasPrefix(objectKey, "../") || strings.Contains(objectKey, "/../") {
		return "", ErrUnsafeObjectKey
	}
	return objectKey, nil
}

func joinObjectURL(baseURL string, objectKey string, expiresAt time.Time) string {
	u, err := url.Parse(baseURL + "/" + pathEscapeObjectKey(objectKey))
	if err != nil {
		return baseURL + "/" + pathEscapeObjectKey(objectKey)
	}
	q := u.Query()
	q.Set("expires", fmt.Sprintf("%d", expiresAt.Unix()))
	u.RawQuery = q.Encode()
	return u.String()
}

func pathEscapeObjectKey(objectKey string) string {
	parts := strings.Split(objectKey, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}
