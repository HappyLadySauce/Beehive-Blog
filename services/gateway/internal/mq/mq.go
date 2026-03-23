package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/code"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/config"
	"github.com/HappyLadySauce/errors"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	exchangeName = "gateway_exchange"
	exchangeType = "direct"

	reconnectDelay = 3 * time.Second
	publishTimeout = 5 * time.Second
)

// Message 是网关间通信的消息结构
type Message struct {
	Action    string                 `json:"action"`
	ConnID    string                 `json:"connId"`
	UserID    string                 `json:"userId"`
	DeviceID  string                 `json:"deviceId"`
	GatewayID string                 `json:"gatewayId"`
	Region    string                 `json:"region,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// MQ RabbitMQ 客户端，支持断线重连
type MQ struct {
	url     string
	mu      sync.RWMutex
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  logx.Logger

	ctx    context.Context
	cancel context.CancelFunc
}

// NewMQ 创建 MQ 客户端并建立连接
func NewMQ(cfg config.Config) (*MQ, error) {
	ctx, cancel := context.WithCancel(context.Background())
	m := &MQ{
		url:    cfg.RabbitMQURL,
		logger: logx.WithContext(ctx),
		ctx:    ctx,
		cancel: cancel,
	}

	if err := m.connect(); err != nil {
		cancel()
		return nil, err
	}

	// 启动后台重连 goroutine
	go m.watchReconnect()

	return m, nil
}

// connect 建立 AMQP 连接并声明 exchange
func (m *MQ) connect() error {
	conn, err := amqp.Dial(m.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// 声明 direct 交换机（幂等，多次声明安全）
	if err := ch.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	m.mu.Lock()
	m.conn = conn
	m.channel = ch
	m.mu.Unlock()

	m.logger.Infof("RabbitMQ connected: %s", m.url)
	return nil
}

// watchReconnect 监听连接关闭事件，自动重连
func (m *MQ) watchReconnect() {
	for {
		m.mu.RLock()
		conn := m.conn
		m.mu.RUnlock()

		if conn == nil {
			return
		}

		// 阻塞等待连接关闭通知
		reason, ok := <-conn.NotifyClose(make(chan *amqp.Error, 1))
		if !ok {
			// 连接被主动关闭（调用了 Close()），退出
			return
		}
		m.logger.Errorf("RabbitMQ connection closed, reason: %v, reconnecting...", reason)

		// 退避重连
		for {
			select {
			case <-m.ctx.Done():
				return
			case <-time.After(reconnectDelay):
			}

			if err := m.connect(); err != nil {
				m.logger.Errorf("RabbitMQ reconnect failed: %v, retrying in %s", err, reconnectDelay)
				continue
			}
			m.logger.Info("RabbitMQ reconnected successfully")
			break
		}
	}
}

// Close 关闭 MQ 客户端
func (m *MQ) Close() error {
	m.cancel() // 通知 watchReconnect 退出

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.channel != nil {
		m.channel.Close()
		m.channel = nil
	}
	if m.conn != nil {
		if err := m.conn.Close(); err != nil {
			return fmt.Errorf("close RabbitMQ connection: %w", err)
		}
		m.conn = nil
	}
	return nil
}

// channel 安全获取当前 channel
func (m *MQ) getChannel() *amqp.Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.channel
}

// Publish 向指定 routingKey 发布消息
func (m *MQ) Publish(routingKey string, msg *Message) error {
	msg.Timestamp = time.Now().UnixMilli()

	data, err := json.Marshal(msg)
	if err != nil {
		return errors.WithCode(code.CodeMarshalFailed, "failed to marshal message")
	}

	ch := m.getChannel()
	if ch == nil {
		return errors.WithCode(code.CodeMQPushFailed, "channel not available, reconnecting")
	}

	ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
	defer cancel()

	if err := ch.PublishWithContext(
		ctx,
		exchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         data,
		},
	); err != nil {
		m.logger.Errorf("failed to publish message to routingKey=%s: %v", routingKey, err)
		return errors.WithCode(code.CodeMQPushFailed, "failed to publish message")
	}

	return nil
}

// SubscribeGateway 为当前网关实例订阅专属队列。
//
// 每个网关实例使用独立队列（queueName 通常为 "gateway.{gatewayID}"），
// 绑定到 exchangeName，routingKey 为 gatewayID，
// 这样 Publish 时指定 routingKey=gatewayID 即可精准投递到目标网关。
//
// 返回的 chan 会在连接断开或 ctx 取消时关闭。
func (m *MQ) SubscribeGateway(ctx context.Context, queueName, routingKey string) (<-chan *Message, error) {
	ch := m.getChannel()
	if ch == nil {
		return nil, errors.WithCode(code.CodeMQConsumeFailed, "channel not available")
	}

	// 声明专属队列（exclusive=false 允许重连后重新消费；auto-delete=true 断开后自动删除）
	if _, err := ch.QueueDeclare(
		queueName,
		false, // durable：网关重启后重新订阅，无需持久化
		true,  // auto-delete：所有消费者断开后自动删除
		false, // exclusive
		false, // no-wait
		nil,
	); err != nil {
		return nil, errors.WithCode(code.CodeMQDeclareFailed, "failed to declare queue: "+queueName)
	}

	if err := ch.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	); err != nil {
		return nil, errors.WithCode(code.CodeMQBindQueueFailed, "failed to bind queue: "+queueName)
	}

	// 每次只处理一条消息，避免消息积压
	if err := ch.Qos(1, 0, false); err != nil {
		return nil, errors.WithCode(code.CodeMQSetQosFailed, "failed to set QoS")
	}

	deliveries, err := ch.Consume(
		queueName,
		"",    // consumer tag，空则自动生成
		false, // auto-ack：手动 ack 保证可靠性
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, errors.WithCode(code.CodeMQConsumeFailed, "failed to start consuming: "+queueName)
	}

	out := make(chan *Message, 64)

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					// channel 被关闭（连接断开），退出等待重连
					m.logger.Errorf("delivery channel closed for queue=%s", queueName)
					return
				}

				var msg Message
				if err := json.Unmarshal(d.Body, &msg); err != nil {
					m.logger.Errorf("unmarshal message failed, queue=%s: %v", queueName, err)
					d.Nack(false, false) // 解析失败，丢弃（避免死循环）
					continue
				}

				select {
				case out <- &msg:
					d.Ack(false)
				case <-ctx.Done():
					d.Nack(false, true) // 重新入队
					return
				default:
					// 消费者处理太慢，消息重新入队
					m.logger.Errorf("message channel full, requeue message connID=%s", msg.ConnID)
					d.Nack(false, true)
				}
			}
		}
	}()

	m.logger.Infof("subscribed queue=%s routingKey=%s", queueName, routingKey)
	return out, nil
}
