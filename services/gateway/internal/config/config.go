// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/ip"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	// Redis 缓存配置
	RedisHost string `json:"RedisHost"`
	RedisPass string `json:"RedisPass"`
	RedisDB   int    `json:"RedisDB"`

	// RabbitMQ 消费配置（用于 message.push）；可选，未配置时不消费 message.created，无实时推送。
	RabbitMQURL      string `json:"RabbitMQURL,optional"`      // amqp://guest:guest@127.0.0.1:5672/
	RabbitMQExchange string `json:"RabbitMQExchange,optional"` // im.events，与 Message 服务发布端一致
	RabbitMQQueue    string `json:"RabbitMQQueue,optional"`    // 每实例独立队列，如 gateway.push.gw-1
	RabbitMQRouteKey string `json:"RabbitMQRouteKey,optional"` // message.created

	// IP 地理位置查询配置
	IPQuery ip.ManagerConfig `json:"IPQuery,optional"`
}
