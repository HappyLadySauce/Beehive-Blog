package ip

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
)

// GeoIP2Querier 使用 MaxMind GeoIP2 数据库的本地查询器
type GeoIP2Querier struct {
	cityDB *geoip2.Reader
	ispDB  *geoip2.Reader
}

// GeoIP2Config GeoIP2 配置
type GeoIP2Config struct {
	CityDBPath string // GeoIP2-City.mmdb 文件路径
	ISPDBPath  string // GeoIP2-ISP.mmdb 文件路径（可选）
}

// NewGeoIP2Querier 创建 GeoIP2 查询器
func NewGeoIP2Querier(cfg GeoIP2Config) (*GeoIP2Querier, error) {
	q := &GeoIP2Querier{}

	// 加载城市数据库（必需）
	if cfg.CityDBPath != "" {
		db, err := geoip2.Open(cfg.CityDBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open city db: %w", err)
		}
		q.cityDB = db
	}

	// 加载 ISP 数据库（可选）
	if cfg.ISPDBPath != "" {
		db, err := geoip2.Open(cfg.ISPDBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open isp db: %w", err)
		}
		q.ispDB = db
	}

	return q, nil
}

// Query 查询IP地址
func (q *GeoIP2Querier) Query(ip string) (*Location, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	loc := &Location{}

	// 查询城市信息
	if q.cityDB != nil {
		record, err := q.cityDB.City(parsedIP)
		if err != nil {
			return nil, fmt.Errorf("city lookup failed: %w", err)
		}

		// 国家名称（优先中文）
		if name, ok := record.Country.Names["zh-CN"]; ok {
			loc.Country = name
		} else if name, ok := record.Country.Names["en"]; ok {
			loc.Country = name
		}
		loc.CountryCode = record.Country.IsoCode

		// 省份/州名称
		if len(record.Subdivisions) > 0 {
			if name, ok := record.Subdivisions[0].Names["zh-CN"]; ok {
				loc.Region = name
			} else if name, ok := record.Subdivisions[0].Names["en"]; ok {
				loc.Region = name
			}
			loc.RegionCode = record.Subdivisions[0].IsoCode
		}

		// 城市名称
		if name, ok := record.City.Names["zh-CN"]; ok {
			loc.City = name
		} else if name, ok := record.City.Names["en"]; ok {
			loc.City = name
		}

		// 经纬度
		loc.Latitude = fmt.Sprintf("%f", record.Location.Latitude)
		loc.Longitude = fmt.Sprintf("%f", record.Location.Longitude)
		loc.Timezone = record.Location.TimeZone
	}

	// 查询 ISP 信息
	if q.ispDB != nil {
		record, err := q.ispDB.ISP(parsedIP)
		if err == nil {
			loc.ISP = record.ISP
		}
	}

	return loc, nil
}

// Close 关闭数据库连接
func (q *GeoIP2Querier) Close() error {
	var errs []error
	if q.cityDB != nil {
		if err := q.cityDB.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if q.ispDB != nil {
		if err := q.ispDB.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close geoip2 databases: %v", errs)
	}
	return nil
}

// Available 检查查询器是否可用
func (q *GeoIP2Querier) Available() bool {
	return q.cityDB != nil
}
