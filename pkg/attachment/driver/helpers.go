package driver

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
)

// ErrUnsafeObjectKey is returned when an object key attempts directory traversal.
// ErrUnsafeObjectKey 在 object key 试图目录穿越时返回。
var ErrUnsafeObjectKey = errors.New("attachment object key is unsafe")

// cleanObjectKey normalizes and validates an object key.
// cleanObjectKey 规范化并校验 object key。
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

// joinObjectURL builds a presigned URL with an expires query parameter.
// joinObjectURL 构造包含 expires 查询参数的预签名 URL。
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

// pathEscapeObjectKey URL-escapes each path segment of the object key independently.
// pathEscapeObjectKey 对 object key 的每个路径段独立进行 URL 转义。
func pathEscapeObjectKey(objectKey string) string {
	parts := strings.Split(objectKey, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}
