package driver

import (
	"io"
	"time"
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
