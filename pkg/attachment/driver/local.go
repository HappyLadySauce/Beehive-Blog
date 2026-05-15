package driver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"time"
)

// LocalDriver stores objects under a configured local root directory.
// LocalDriver 将对象存储在配置的本地根目录下。
type LocalDriver struct {
	root string
}

// NewLocalDriver creates a LocalDriver from a JSON config blob.
// Expected config: {"root": "data/attachments"}.
// NewLocalDriver 从 JSON 配置创建 LocalDriver。
// 配置格式：{"root": "data/attachments"}。
func NewLocalDriver(config json.RawMessage) (DriverBackend, error) {
	var cfg struct {
		Root string `json:"root"`
	}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("local driver config: %w", err)
	}
	cfg.Root = strings.TrimSpace(cfg.Root)
	if cfg.Root == "" {
		return nil, fmt.Errorf("local driver: root is required")
	}
	abs, err := filepath.Abs(cfg.Root)
	if err != nil {
		return nil, fmt.Errorf("local driver root: %w", err)
	}
	return &LocalDriver{root: abs}, nil
}

func (d *LocalDriver) DriverName() string {
	return DriverLocal
}

// Save writes the request body to disk under the local root.
// Save 将请求体写入本地存储根目录下的文件。
func (d *LocalDriver) Save(ctx context.Context, req PutRequest) (StoredObject, error) {
	if req.Reader == nil {
		return StoredObject{}, fmt.Errorf("attachment reader is nil")
	}
	localPath, err := cleanObjectKey(req.ObjectKey)
	if err != nil {
		return StoredObject{}, err
	}
	abs, err := d.LocalFilePath(localPath)
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

func (d *LocalDriver) PresignUpload(context.Context, PresignRequest) (PresignResult, error) {
	return PresignResult{}, ErrUnsupportedDriver
}

func (d *LocalDriver) PresignDownload(context.Context, string, time.Duration) (PresignResult, error) {
	return PresignResult{}, ErrUnsupportedDriver
}

// LocalFilePath resolves a relative path against the local root with traversal protection.
// LocalFilePath 将相对路径解析为本地根目录下的绝对路径，并有防目录穿越保护。
func (d *LocalDriver) LocalFilePath(localPath string) (string, error) {
	clean, err := cleanObjectKey(localPath)
	if err != nil {
		return "", err
	}
	abs := filepath.Join(d.root, filepath.FromSlash(clean))
	rel, err := filepath.Rel(d.root, abs)
	if err != nil {
		return "", fmt.Errorf("resolve attachment path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", ErrUnsafeObjectKey
	}
	return abs, nil
}

// Delete removes a local file. Returns nil if the file does not exist.
// Delete 删除本地文件。文件不存在时不报错。
func (d *LocalDriver) Delete(ctx context.Context, objectKey string) error {
	path, err := d.LocalFilePath(objectKey)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete local object: %w", err)
	}
	return nil
}

// HealthCheck verifies the root directory is writable.
// HealthCheck 验证根目录可写。
func (d *LocalDriver) HealthCheck(ctx context.Context) error {
	if err := os.MkdirAll(d.root, 0o755); err != nil {
		return fmt.Errorf("local driver root: %w", err)
	}
	testFile := filepath.Join(d.root, ".healthcheck")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("local driver health check: %w", err)
	}
	f.Close()
	_ = os.Remove(testFile)
	return nil
}
