package ip

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 默认超时时间
	defaultTimeout = 5 * time.Second
)

// Location IP地理位置信息
type Location struct {
	IP          string `json:"ip"`          // IP地址
	Country     string `json:"country"`     // 国家
	CountryCode string `json:"countryCode"` // 国家代码
	Region      string `json:"region"`      // 省份/州
	RegionCode  string `json:"regionCode"`  // 省份/州代码
	City        string `json:"city"`        // 城市
	District    string `json:"district"`    // 区县
	ISP         string `json:"isp"`         // 运营商
	Latitude    string `json:"lat"`         // 纬度
	Longitude   string `json:"lng"`         // 经度
	Timezone    string `json:"timezone"`    // 时区
}

// String 返回格式化的地址字符串
func (l *Location) String() string {
	if l == nil {
		return ""
	}
	parts := []string{}
	if l.Country != "" {
		parts = append(parts, l.Country)
	}
	if l.Region != "" {
		parts = append(parts, l.Region)
	}
	if l.City != "" {
		parts = append(parts, l.City)
	}
	if l.District != "" {
		parts = append(parts, l.District)
	}
	if l.ISP != "" {
		parts = append(parts, l.ISP)
	}
	return strings.Join(parts, " ")
}

// SimpleLocation 返回简化的地址（省市区）
func (l *Location) SimpleLocation() string {
	if l == nil {
		return ""
	}
	parts := []string{}
	if l.Region != "" {
		parts = append(parts, l.Region)
	}
	if l.City != "" && l.City != l.Region {
		parts = append(parts, l.City)
	}
	if l.District != "" {
		parts = append(parts, l.District)
	}
	return strings.Join(parts, "")
}

// Querier IP查询接口
type Querier interface {
	Query(ip string) (*Location, error)
}

// Client IP查询客户端
type Client struct {
	queriers []Querier
	timeout  time.Duration
}

// NewClient 创建IP查询客户端
func NewClient(opts ...Option) *Client {
	c := &Client{
		queriers: []Querier{},
		timeout:  defaultTimeout,
	}
	for _, opt := range opts {
		opt(c)
	}
	// 默认添加本地查询
	if len(c.queriers) == 0 {
		c.queriers = append(c.queriers, NewLocalQuerier())
	}
	return c
}

// Option 客户端配置选项
type Option func(*Client)

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithQuerier 添加查询器
func WithQuerier(q Querier) Option {
	return func(c *Client) {
		c.queriers = append(c.queriers, q)
	}
}

// Query 查询IP地址信息，按顺序尝试所有查询器
func (c *Client) Query(ip string) (*Location, error) {
	// 清理IP地址（去除端口等）
	ip = CleanIP(ip)

	// 验证IP地址
	if !IsValidIP(ip) {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	// 本地地址直接返回
	if IsPrivateIP(ip) {
		return &Location{
			IP:      ip,
			Country: "本地",
			ISP:     "内网",
		}, nil
	}

	// 依次尝试查询器
	for _, q := range c.queriers {
		loc, err := q.Query(ip)
		if err != nil {
			logx.Errorf("IP query failed with %T: %v", q, err)
			continue
		}
		if loc != nil {
			loc.IP = ip
			return loc, nil
		}
	}

	return nil, fmt.Errorf("all queriers failed for IP: %s", ip)
}

// CleanIP 清理IP地址，去除端口等
func CleanIP(ip string) string {
	// 处理 IPv6 带端口的情况 [::1]:8080
	if strings.HasPrefix(ip, "[") {
		if idx := strings.LastIndex(ip, "]"); idx != -1 {
			ip = ip[1:idx]
		}
	} else {
		// 处理 IPv4 带端口的情况 127.0.0.1:8080
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			// 检查是否是IPv6地址（包含多个冒号）
			if strings.Count(ip, ":") == 1 {
				ip = ip[:idx]
			}
		}
	}
	return ip
}

// IsValidIP 验证IP地址是否有效
func IsValidIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}

// IsPrivateIP 判断是否是内网IP
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// IPv4 内网地址段
	privateRanges := []string{
		"10.0.0.0/8",     // 10.0.0.0 - 10.255.255.255
		"172.16.0.0/12",  // 172.16.0.0 - 172.31.255.255
		"192.168.0.0/16", // 192.168.0.0 - 192.168.255.255
		"127.0.0.0/8",    // 127.0.0.0 - 127.255.255.255
		"169.254.0.0/16", // 169.254.0.0 - 169.254.255.255 (链路本地)
		"::1/128",        // IPv6 本地回环
		"fe80::/10",      // IPv6 链路本地
		"fc00::/7",       // IPv6 私有地址
	}

	for _, cidr := range privateRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(parsedIP) {
			return true
		}
	}
	return false
}

// GetClientIP 从HTTP请求中获取客户端IP
func GetClientIP(r *http.Request) string {
	// 优先从 X-Forwarded-For 获取
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For 可能包含多个IP，取第一个
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return CleanIP(strings.TrimSpace(ips[0]))
		}
	}

	// 从 X-Real-IP 获取
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return CleanIP(xri)
	}

	// 从 RemoteAddr 获取
	return CleanIP(r.RemoteAddr)
}

// LocalQuerier 本地查询器（仅返回基础信息）
type LocalQuerier struct{}

// NewLocalQuerier 创建本地查询器
func NewLocalQuerier() *LocalQuerier {
	return &LocalQuerier{}
}

// Query 本地查询
func (l *LocalQuerier) Query(ip string) (*Location, error) {
	return &Location{
		IP:      ip,
		Country: "未知",
		ISP:     "未知",
	}, nil
}
