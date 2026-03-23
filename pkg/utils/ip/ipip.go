package ip

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// IPIPQuerier IPIP.net 查询器（国内精准，需要token）
type IPIPQuerier struct {
	client *http.Client
	token  string
}

// IPIPResponse IPIP.net 响应结构
type IPIPResponse struct {
	IP       string   `json:"ip"`
	Location []string `json:"location"` // [国家, 省份, 城市, 区县, 运营商]
}

// NewIPIPQuerier 创建 IPIP.net 查询器
// token: 从 https://www.ipip.net/ 获取的API Token
func NewIPIPQuerier(token string) *IPIPQuerier {
	return &IPIPQuerier{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		token: token,
	}
}

// Query 查询IP地址
func (q *IPIPQuerier) Query(ip string) (*Location, error) {
	url := fmt.Sprintf("https://api.ipip.net/v1/ip?ip=%s&token=%s", ip, q.token)
	
	resp, err := q.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ipip returned status: %d", resp.StatusCode)
	}

	var result IPIPResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	loc := &Location{}
	if len(result.Location) > 0 {
		loc.Country = result.Location[0]
	}
	if len(result.Location) > 1 {
		loc.Region = result.Location[1]
	}
	if len(result.Location) > 2 {
		loc.City = result.Location[2]
	}
	if len(result.Location) > 3 {
		loc.District = result.Location[3]
	}
	if len(result.Location) > 4 {
		loc.ISP = result.Location[4]
	}

	return loc, nil
}
