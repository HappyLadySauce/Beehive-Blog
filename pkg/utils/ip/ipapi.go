package ip

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// IPAPIQuerier ip-api.com 查询器（免费，支持国内外）
// 注意：免费版有速率限制，商业使用请购买付费版
type IPAPIQuerier struct {
	client  *http.Client
	apiKey  string // 付费版API密钥（可选）
	lang    string // 返回语言
}

// IPAPIResponse ip-api.com 响应结构
type IPAPIResponse struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"regionName"`
	RegionCode  string  `json:"region"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
	Message     string  `json:"message"` // 错误信息
}

// NewIPAPIQuerier 创建 ip-api.com 查询器
// lang: 语言代码，如 "zh-CN", "en" 等
func NewIPAPIQuerier(lang string) *IPAPIQuerier {
	if lang == "" {
		lang = "zh-CN"
	}
	return &IPAPIQuerier{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		lang: lang,
	}
}

// NewIPAPIQuerierWithKey 创建带API密钥的 ip-api.com 查询器（付费版）
func NewIPAPIQuerierWithKey(apiKey, lang string) *IPAPIQuerier {
	q := NewIPAPIQuerier(lang)
	q.apiKey = apiKey
	return q
}

// Query 查询IP地址
func (q *IPAPIQuerier) Query(ip string) (*Location, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s?lang=%s", ip, q.lang)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 付费版需要添加API密钥
	if q.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+q.apiKey)
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ip-api returned status: %d", resp.StatusCode)
	}

	var result IPAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("ip-api error: %s", result.Message)
	}

	return &Location{
		Country:     result.Country,
		CountryCode: result.CountryCode,
		Region:      result.Region,
		RegionCode:  result.RegionCode,
		City:        result.City,
		ISP:         result.ISP,
		Latitude:    fmt.Sprintf("%f", result.Lat),
		Longitude:   fmt.Sprintf("%f", result.Lon),
		Timezone:    result.Timezone,
	}, nil
}
