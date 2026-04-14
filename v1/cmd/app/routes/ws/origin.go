package ws

import (
	"net/url"
	"strings"
)

// checkOrigin 与 middlewares.Cors 的 AllowOriginFunc 语义对齐，供 WebSocket Upgrade 校验。
func checkOrigin(origin string) bool {
	if origin == "" {
		return true
	}
	if origin == "https://github.com" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	return false
}
