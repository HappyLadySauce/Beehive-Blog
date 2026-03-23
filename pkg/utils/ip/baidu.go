package ip

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// BaiduQuerier 百度地图IP定位API查询器
type BaiduQuerier struct {
	client *http.Client
	ak     string // 百度地图API密钥
}

// BaiduResponse 百度地图API响应结构
type BaiduResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Content struct {
		Address       string `json:"address"`
		AddressDetail struct {
			Country   string `json:"country"`
			Province  string `json:"province"`
			City      string `json:"city"`
			District  string `json:"district"`
			Street    string `json:"street"`
			StreetNum string `json:"street_number"`
			CityCode  int    `json:"city_code"`
		} `json:"address_detail"`
		Point struct {
			X string `json:"x"` // 经度
			Y string `json:"y"` // 纬度
		} `json:"point"`
	} `json:"content"`
}

// NewBaiduQuerier 创建百度地图IP定位查询器
// ak: 百度地图开放平台申请的API Key
// 文档: https://lbsyun.baidu.com/index.php?title=webapi/ip-api
func NewBaiduQuerier(ak string) *BaiduQuerier {
	return &BaiduQuerier{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		ak: ak,
	}
}

// Query 查询IP地址
func (q *BaiduQuerier) Query(ip string) (*Location, error) {
	// 百度API不支持IPv6，如果是IPv6直接返回错误
	if IsIPv6(ip) {
		return nil, fmt.Errorf("baidu api does not support IPv6")
	}

	url := fmt.Sprintf("https://api.map.baidu.com/location/ip?ak=%s&ip=%s&coor=bd09ll", q.ak, ip)

	resp, err := q.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("baidu api returned status: %d", resp.StatusCode)
	}

	var result BaiduResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != 0 {
		return nil, fmt.Errorf("baidu api error: %s", result.Message)
	}

	detail := result.Content.AddressDetail
	return &Location{
		Country:   detail.Country,
		Region:    detail.Province,
		City:      detail.City,
		District:  detail.District,
		Latitude:  result.Content.Point.Y,
		Longitude: result.Content.Point.X,
	}, nil
}

// IsIPv6 判断是否是IPv6地址
func IsIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.To4() == nil && parsedIP.To16() != nil
}
