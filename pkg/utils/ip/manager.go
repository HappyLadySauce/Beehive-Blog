package ip

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// QuerySource 查询源类型
type QuerySource string

const (
	SourceGeoIP2 QuerySource = "geoip2" // 本地 GeoIP2 数据库（优先级最高）
	SourceIPIP   QuerySource = "ipip"   // IPIP.net 在线API
	SourceBaidu  QuerySource = "baidu"  // 百度地图API
	SourceIPAPI  QuerySource = "ipapi"  // ip-api.com 在线API（免费，优先级最低）
)

// ManagerConfig 查询管理器配置
type ManagerConfig struct {
	// 超时时间
	Timeout int `json:"Timeout,optional"`

	// GeoIP2 配置
	GeoIP2Enabled bool   `json:"GeoIP2Enabled,optional"` // 是否启用 GeoIP2
	GeoIP2CityDB  string `json:"GeoIP2CityDB,optional"`  // City 数据库路径
	GeoIP2ISPDB   string `json:"GeoIP2ISPDB,optional"`   // ISP 数据库路径（可选）

	// GeoLite2 自动下载配置
	GeoLite2AutoDownload bool   `json:"GeoLite2AutoDownload,optional"` // 是否自动下载 GeoLite2
	GeoLite2LicenseKey   string `json:"GeoLite2LicenseKey,optional"`   // MaxMind License Key（建议从环境变量 GEO_LITE2_LICENSE_KEY 读取）
	GeoLite2DataDir      string `json:"GeoLite2DataDir,optional"`      // 数据存放目录，默认 ./data

	// IPIP.net 配置
	IPIPEnabled bool   `json:"IPIPEnabled,optional"` // 是否启用 IPIP
	IPIPToken   string `json:"IPIPToken,optional"`   // IPIP Token（建议从环境变量 IPIP_TOKEN 读取）

	// 百度地图配置
	BaiduEnabled bool   `json:"BaiduEnabled,optional"` // 是否启用百度
	BaiduAK      string `json:"BaiduAK,optional"`      // 百度 AK（建议从环境变量 BAIDU_AK 读取）

	// ip-api.com 配置
	IPAPIEnabled bool   `json:"IPAPIEnabled,optional"` // 是否启用 IPAPI
	IPAPILang    string `json:"IPAPILang,optional"`    // 语言，默认 zh-CN
}

// loadFromEnv 从环境变量加载敏感配置
func (c *ManagerConfig) loadFromEnv() {
	// 如果配置为空，尝试从环境变量读取
	if c.GeoLite2LicenseKey == "" {
		c.GeoLite2LicenseKey = os.Getenv("GEO_LITE2_LICENSE_KEY")
	}
	if c.IPIPToken == "" {
		c.IPIPToken = os.Getenv("IPIP_TOKEN")
	}
	if c.BaiduAK == "" {
		c.BaiduAK = os.Getenv("BAIDU_AK")
	}
}

// Manager IP查询管理器
type Manager struct {
	config     ManagerConfig
	queriers   map[QuerySource]Querier
	priority   []QuerySource
	mu         sync.RWMutex
	timeout    time.Duration
	geoLite2DL *GeoLite2Downloader // GeoLite2 下载器
}

var (
	managerInstance *Manager
	managerOnce     sync.Once
)

// NewManager 创建查询管理器
func NewManager(cfg ManagerConfig) (*Manager, error) {
	// 从环境变量加载敏感配置
	cfg.loadFromEnv()

	m := &Manager{
		config:   cfg,
		queriers: make(map[QuerySource]Querier),
		priority: make([]QuerySource, 0),
	}

	// 设置超时
	if cfg.Timeout > 0 {
		m.timeout = time.Duration(cfg.Timeout) * time.Second
	} else {
		m.timeout = 5 * time.Second
	}

	// 自动下载 GeoLite2 数据库
	if cfg.GeoLite2AutoDownload && cfg.GeoLite2LicenseKey != "" {
		dl := NewGeoLite2Downloader(GeoLite2Config{
			LicenseKey: cfg.GeoLite2LicenseKey,
			DataDir:    cfg.GeoLite2DataDir,
		})
		m.geoLite2DL = dl

		// 首次下载
		logx.Info("Downloading GeoLite2 databases...")
		if err := dl.Download(); err != nil {
			logx.Errorf("Failed to download GeoLite2 databases: %v", err)
			// 下载失败不阻塞，继续尝试使用已有文件或 fallback
		}

		// 启动自动更新
		dl.StartAutoUpdate()

		// 更新配置中的数据库路径
		cityDB, ispDB := dl.GetDBPaths()
		if cfg.GeoIP2CityDB == "" && cityDB != "" {
			cfg.GeoIP2CityDB = cityDB
		}
		if cfg.GeoIP2ISPDB == "" && ispDB != "" {
			cfg.GeoIP2ISPDB = ispDB
		}
	}

	// 按优先级初始化查询器

	// 1. GeoIP2（本地数据库，优先级最高）
	if cfg.GeoIP2Enabled && cfg.GeoIP2CityDB != "" {
		geoip2Querier, err := NewGeoIP2Querier(GeoIP2Config{
			CityDBPath: cfg.GeoIP2CityDB,
			ISPDBPath:  cfg.GeoIP2ISPDB,
		})
		if err != nil {
			logx.Errorf("Failed to initialize GeoIP2: %v", err)
		} else {
			m.queriers[SourceGeoIP2] = geoip2Querier
			m.priority = append(m.priority, SourceGeoIP2)
			logx.Info("GeoIP2 querier initialized")
		}
	}

	// 2. IPIP.net（国内精准）
	if cfg.IPIPEnabled && cfg.IPIPToken != "" {
		m.queriers[SourceIPIP] = NewIPIPQuerier(cfg.IPIPToken)
		m.priority = append(m.priority, SourceIPIP)
		logx.Info("IPIP querier initialized")
	}

	// 3. 百度地图（国内）
	if cfg.BaiduEnabled && cfg.BaiduAK != "" {
		m.queriers[SourceBaidu] = NewBaiduQuerier(cfg.BaiduAK)
		m.priority = append(m.priority, SourceBaidu)
		logx.Info("Baidu querier initialized")
	}

	// 4. ip-api.com（免费，国内外，优先级最低）
	if cfg.IPAPIEnabled {
		lang := cfg.IPAPILang
		if lang == "" {
			lang = "zh-CN"
		}
		m.queriers[SourceIPAPI] = NewIPAPIQuerier(lang)
		m.priority = append(m.priority, SourceIPAPI)
		logx.Info("IPAPI querier initialized")
	}

	// 如果没有配置任何查询器，添加本地查询器作为 fallback
	if len(m.priority) == 0 {
		m.queriers[SourceIPAPI] = NewLocalQuerier()
		m.priority = append(m.priority, SourceIPAPI)
		logx.Error("No IP querier configured, using local querier")
	}

	logx.Infof("IP Manager initialized with %d queriers, priority: %v", len(m.priority), m.priority)
	return m, nil
}

// InitManager 初始化全局管理器（单例）
func InitManager(cfg ManagerConfig) error {
	var err error
	managerOnce.Do(func() {
		managerInstance, err = NewManager(cfg)
	})
	return err
}

// GetManager 获取全局管理器实例
func GetManager() *Manager {
	return managerInstance
}

// Query 查询IP地址（按优先级依次尝试）
func (m *Manager) Query(ip string) (*Location, error) {
	if m == nil {
		return nil, fmt.Errorf("IP manager not initialized")
	}

	// 清理和验证IP
	ip = CleanIP(ip)
	if !IsValidIP(ip) {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	// 内网IP直接返回
	if IsPrivateIP(ip) {
		return &Location{
			IP:      ip,
			Country: "本地",
			ISP:     "内网",
		}, nil
	}

	m.mu.RLock()
	priority := make([]QuerySource, len(m.priority))
	copy(priority, m.priority)
	queriers := m.queriers
	m.mu.RUnlock()

	// 按优先级依次查询
	for _, source := range priority {
		querier, ok := queriers[source]
		if !ok {
			continue
		}

		// 使用带超时的查询
		loc, err := m.queryWithTimeout(querier, ip)
		if err != nil {
			logx.Errorf("IP query failed with %s: %v", source, err)
			continue
		}

		if loc != nil {
			loc.IP = ip
			logx.Debugf("IP %s query success with %s: %s", ip, source, loc.String())
			return loc, nil
		}
	}

	return nil, fmt.Errorf("all queriers failed for IP: %s", ip)
}

// queryWithTimeout 带超时的查询
func (m *Manager) queryWithTimeout(q Querier, ip string) (*Location, error) {
	type result struct {
		loc *Location
		err error
	}

	ch := make(chan result, 1)
	go func() {
		loc, err := q.Query(ip)
		ch <- result{loc, err}
	}()

	select {
	case <-time.After(m.timeout):
		return nil, fmt.Errorf("query timeout")
	case r := <-ch:
		return r.loc, r.err
	}
}

// QueryWithSource 使用指定查询源查询
func (m *Manager) QueryWithSource(ip string, source QuerySource) (*Location, error) {
	if m == nil {
		return nil, fmt.Errorf("IP manager not initialized")
	}

	m.mu.RLock()
	querier, ok := m.queriers[source]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("querier %s not available", source)
	}

	ip = CleanIP(ip)
	if !IsValidIP(ip) {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	loc, err := querier.Query(ip)
	if err != nil {
		return nil, err
	}
	if loc != nil {
		loc.IP = ip
	}
	return loc, nil
}

// GetAvailableSources 获取可用的查询源列表
func (m *Manager) GetAvailableSources() []QuerySource {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	sources := make([]QuerySource, len(m.priority))
	copy(sources, m.priority)
	return sources
}

// Close 关闭管理器，释放资源
func (m *Manager) Close() error {
	if m == nil {
		return nil
	}

	// 停止 GeoLite2 自动更新
	if m.geoLite2DL != nil {
		m.geoLite2DL.StopAutoUpdate()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for source, querier := range m.queriers {
		// 如果查询器实现了 Close 方法，调用它
		if closer, ok := querier.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close %s: %w", source, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing queriers: %v", errs)
	}
	return nil
}

// MustQuery 查询IP地址，失败时返回空对象而非错误
func (m *Manager) MustQuery(ip string) *Location {
	loc, err := m.Query(ip)
	if err != nil {
		logx.Errorf("IP query failed for %s: %v", ip, err)
		return &Location{
			IP:      CleanIP(ip),
			Country: "未知",
			ISP:     "未知",
		}
	}
	return loc
}
