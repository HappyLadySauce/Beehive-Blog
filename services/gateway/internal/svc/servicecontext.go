// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"encoding/json"
	"fmt"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/ip"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/config"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/mq"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/ws"
	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config    config.Config
	Hub       *ws.Hub
	MQ        *mq.MQ
	IPManager *ip.Manager
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 创建连接管理器
	hub, err := ws.NewHub("", c)
	if err != nil {
		panic(fmt.Errorf("create hub: %w", err))
	}

	// 创建 MQ 客户端（RabbitMQ 未配置时为 nil，后续优雅降级）
	var messageQueue *mq.MQ
	if c.RabbitMQURL != "" {
		messageQueue, err = mq.NewMQ(c)
		if err != nil {
			panic(fmt.Errorf("create mq: %w", err))
		}
	} else {
		logx.Info("RabbitMQ not configured, cross-gateway sync disabled")
	}

	// 注入跨网关关闭回调：当旧连接在其他网关时，通过 MQ 发送 close 指令
	if messageQueue != nil {
		hub.OnRemoteClose = func(info ws.ConnectionInfo) {
			msg := &mq.Message{
				Action:    "close",
				ConnID:    info.ConnID,
				UserID:    info.UserID,
				DeviceID:  info.DeviceID,
				GatewayID: info.GatewayID,
				Region:    info.Region,
			}
			if err := messageQueue.Publish(info.GatewayID, msg); err != nil {
				logx.Errorf("publish close msg to gateway %s failed: %v", info.GatewayID, err)
			}
		}
	}

	// 初始化 IP 查询管理器
	var ipManager *ip.Manager
	if c.IPQuery.GeoIP2Enabled || c.IPQuery.IPIPEnabled || c.IPQuery.BaiduEnabled || c.IPQuery.IPAPIEnabled {
		var err error
		ipManager, err = ip.NewManager(c.IPQuery)
		if err != nil {
			logx.Errorf("Failed to initialize IP manager: %v", err)
			// 不 panic，允许服务继续启动
		}
	} else {
		logx.Info("IP query not configured")
	}

	svc := &ServiceContext{
		Config:    c,
		Hub:       hub,
		MQ:        messageQueue,
		IPManager: ipManager,
	}

	// 启动 MQ 订阅，监听发给本网关的消息
	if messageQueue != nil {
		go svc.startGatewayConsumer()
	}

	return svc
}

// startGatewayConsumer 订阅本网关专属队列，处理来自其他网关的消息
func (s *ServiceContext) startGatewayConsumer() {
	// 使用 Hub 的 context，Hub 关闭时 consumer 也会退出
	ctx := s.Hub.Context()
	queueName := "gateway." + s.Hub.GatewayID()
	routingKey := s.Hub.GatewayID()

	msgs, err := s.MQ.SubscribeGateway(ctx, queueName, routingKey)
	if err != nil {
		logx.Errorf("subscribe gateway queue failed: %v", err)
		return
	}

	logx.Infof("gateway consumer started, queue=%s, routingKey=%s", queueName, routingKey)

	for msg := range msgs {
		switch msg.Action {
		case "close":
			// 收到关闭指令，踢掉本地连接
			conn := s.Hub.GetConn(msg.ConnID)
			if conn != nil {
				if err := conn.Close(); err != nil {
					logx.Errorf("close conn %s by mq failed: %v", msg.ConnID, err)
				} else {
					logx.Infof("closed conn %s by mq (region: %s)", msg.ConnID, msg.Region)
				}
			} else {
				logx.Infof("conn %s not found in local hub, skip close", msg.ConnID)
			}

		case "push":
			// 向本地连接推送消息
			if len(msg.Data) > 0 {
				payload, _ := json.Marshal(msg.Data)
				if err := s.Hub.SendToConn(msg.ConnID, payload); err != nil {
					logx.Errorf("push to conn %s failed: %v", msg.ConnID, err)
				} else {
					logx.Infof("pushed msg to conn %s", msg.ConnID)
				}
			}

		default:
			logx.Infof("unknown mq action: %s, connID=%s", msg.Action, msg.ConnID)
		}
	}

	logx.Info("gateway consumer stopped")
}

func (s *ServiceContext) Close() {
	if s.Hub != nil {
		if err := s.Hub.Close(); err != nil {
			logx.Errorf("hub close error: %v", err)
		}
	}
	if s.MQ != nil {
		if err := s.MQ.Close(); err != nil {
			logx.Errorf("mq close error: %v", err)
		}
	}
}
