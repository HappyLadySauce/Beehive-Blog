package ip

import (
	"fmt"
	"log"
	"net/http"
)

// ExampleClient_Query 示例：基本查询
func ExampleClient_Query() {
	// 创建客户端，使用多个查询器（按优先级排序）
	client := NewClient(
		WithTimeout(5),
		// 使用 IPIP.net（国内精准，需要token）
		// WithQuerier(NewIPIPQuerier("your-ipip-token")),

		// 使用百度地图（国内，需要ak）
		// WithQuerier(NewBaiduQuerier("your-baidu-ak")),

		// 使用 ip-api.com（免费，国内外都支持）
		WithQuerier(NewIPAPIQuerier("zh-CN")),
	)

	// 查询IP
	loc, err := client.Query("8.8.8.8")
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	fmt.Printf("IP: %s\n", loc.IP)
	fmt.Printf("国家: %s\n", loc.Country)
	fmt.Printf("省份: %s\n", loc.Region)
	fmt.Printf("城市: %s\n", loc.City)
	fmt.Printf("运营商: %s\n", loc.ISP)
	fmt.Printf("完整地址: %s\n", loc.String())
	fmt.Printf("简化地址: %s\n", loc.SimpleLocation())
}

// ExampleGetClientIP 示例：从HTTP请求获取客户端IP
func ExampleGetClientIP() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取客户端IP（自动处理 X-Forwarded-For 等头部）
		clientIP := GetClientIP(r)

		fmt.Printf("Client IP: %s\n", clientIP)

		// 查询IP地址
		client := NewClient(WithQuerier(NewIPAPIQuerier("zh-CN")))
		loc, err := client.Query(clientIP)
		if err != nil {
			fmt.Printf("Query failed: %v\n", err)
			return
		}

		fmt.Printf("Location: %s\n", loc.String())
	})

	_ = handler
}

// ExampleIsPrivateIP 示例：判断内网IP
func ExampleIsPrivateIP() {
	ips := []string{
		"192.168.1.1",
		"10.0.0.1",
		"127.0.0.1",
		"8.8.8.8",
		"::1",
	}

	for _, ip := range ips {
		if IsPrivateIP(ip) {
			fmt.Printf("%s 是内网IP\n", ip)
		} else {
			fmt.Printf("%s 是公网IP\n", ip)
		}
	}
}
