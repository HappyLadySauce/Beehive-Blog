package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/code"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/config"

	"github.com/HappyLadySauce/errors"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	lockKeyPrefix    = "ws:lock:"
	connKeyPrefix    = "ws:conn:"
	userDevKeyPrefix = "ws:u:d:"

	connTTL = 24 * time.Hour
)

// Connection 表示单条 WebSocket 连接
type Connection struct {
	ConnID    string
	UserID    string
	DeviceID  string
	GatewayID string
	IP        string
	Region    string
	conn      *websocket.Conn
	writeMu   sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

// Context 返回连接的 context，用于监听连接关闭事件
func (c *Connection) Context() context.Context {
	return c.ctx
}

// RawConn 返回底层的 WebSocket 连接，用于设置读写参数
// 注意：不要直接操作 RawConn，应该通过 Connection 的方法
func (c *Connection) RawConn() *websocket.Conn {
	return c.conn
}

// ConnectionInfo 是存储在 Redis 中的连接元数据
type ConnectionInfo struct {
	ConnID    string `json:"connId"`
	UserID    string `json:"userId"`
	DeviceID  string `json:"deviceId"`
	GatewayID string `json:"gatewayId"`
	IP        string `json:"ip"`
	Region    string `json:"region"`
	Status    string `json:"status"` // "online" | "closing"
}

// Hub 管理当前进程所有 WebSocket 连接
type Hub struct {
	gatewayID string
	conns     map[string]*Connection
	mu        sync.RWMutex
	rdb       *redis.Client // 修复：使用指针类型
	rs        *redsync.Redsync
	logx.Logger

	ctx    context.Context
	cancel context.CancelFunc

	// 跨网关关闭回调：当发现旧连接在其他网关时调用，由外部注入
	// 参数：旧连接信息；实现者负责通过 MQ 发送 close 指令
	OnRemoteClose func(info ConnectionInfo)
}

// NewHub 创建连接管理器，gatewayID 为空时自动生成。
func NewHub(gatewayID string, cfg config.Config) (*Hub, error) {
	if gatewayID == "" {
		gatewayID = "gateway-" + uuid.Must(uuid.NewRandom()).String()
	}

	ctx, cancel := context.WithCancel(context.Background())
	registered := false
	defer func() {
		if !registered {
			cancel()
		}
	}()

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	registered = true

	return &Hub{
		ctx:       ctx,
		cancel:    cancel,
		gatewayID: gatewayID,
		conns:     make(map[string]*Connection),
		rdb:       client, // 修复：赋值 Redis 客户端
		rs:        rs,
		Logger:    logx.WithContext(ctx),
	}, nil
}

// GatewayID 返回当前网关 ID（供 MQ 订阅时使用）
func (h *Hub) GatewayID() string {
	return h.gatewayID
}

// Context 返回 Hub 的 context，用于监听 Hub 关闭事件
func (h *Hub) Context() context.Context {
	return h.ctx
}

// Close 关闭 Hub，断开所有连接并释放 Redis 客户端
// 返回所有关闭过程中的错误（使用 Go 1.20+ 的 errors.Join 合并）
func (h *Hub) Close() error {
	h.cancel()

	h.mu.Lock()
	defer h.mu.Unlock()

	var errs error
	for connID, conn := range h.conns {
		if err := conn.closeConn(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("close conn %s: %w", connID, err))
		}
		delete(h.conns, connID)
	}

	if h.rdb != nil {
		if err := h.rdb.Close(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("close redis: %w", err))
		}
	}

	if errs == nil {
		return nil
	}
	// errors.Join 将多个错误合并为一个
	// 使用 errors.Is / errors.As 可以逐个检查
	return errs
}

// closeConn 关闭底层 WebSocket（内部使用，调用方需持有 writeMu 或确保安全）
func (c *Connection) closeConn() error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.conn != nil {
		// 发送 CloseFrame 给客户端
		_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// Close 关闭连接（对外暴露）
// 会取消 context、关闭底层 WebSocket，通知所有监听该连接的 goroutine 退出
func (c *Connection) Close() error {
	// 先取消 context，通知所有依赖此连接的 goroutine 退出
	c.cancel()

	info := fmt.Sprintf("connID=%s userID=%s deviceID=%s gatewayID=%s",
		c.ConnID, c.UserID, c.DeviceID, c.GatewayID)
	if err := c.closeConn(); err != nil {
		return errors.Wrapf(err, "failed to close WebSocket connection: %s", info)
	}
	return nil
}

// connKey 生成 Redis 连接 key，使用 ":" 分隔避免 userID/deviceID 拼接碰撞
func connKey(userID, deviceID string) string {
	return connKeyPrefix + userID + ":" + deviceID
}

// userDevKey 生成 user-device 映射 key
func userDevKey(userID, deviceID string) string {
	return userDevKeyPrefix + userID + ":" + deviceID
}

// Register 将已升级的 WebSocket 注册到 Hub，返回带 ConnID 的 Connection。
// 若同一 user+device 已有旧连接：
//   - 旧连接在本网关 → 直接关闭
//   - 旧连接在其他网关 → 触发 OnRemoteClose 回调（由调用方通过 MQ 通知）
//
// location 参数为可选的地理位置信息（如 "中国 广东 深圳 电信"）
func (h *Hub) Register(conn *websocket.Conn, userID, deviceID, ip, location string) (*Connection, error) {
	// connID 格式：userID:deviceID，与 Redis key 保持一致
	connID := userID + ":" + deviceID
	redisConnKey := connKey(userID, deviceID)
	redisUserDevKey := userDevKey(userID, deviceID)

	// 获取分布式锁，防止同一 connID 并发注册
	mutex := h.rs.NewMutex(lockKeyPrefix + connID)
	if err := mutex.Lock(); err != nil {
		h.Logger.Errorf("failed to acquire lock for connID=%s: %v", connID, err)
		return nil, errors.WithCode(code.CodeDistributedLockGetFailed, "failed to acquire distributed lock")
	}
	defer mutex.Unlock()

	// 查询 Redis 中是否存在旧连接
	val, err := h.rdb.Get(h.ctx, redisConnKey).Result()
	if err != nil && err != redis.Nil {
		h.Logger.Errorf("get connectionInfo from redis failed, connID=%s: %v", connID, err)
		return nil, errors.WithCode(code.CodeCacheGetFailed, "get connectionInfo from redis failed")
	}

	if err == nil && val != "" {
		// 修复：err == nil 表示成功读取，才进行 Unmarshal
		var oldConnInfo ConnectionInfo
		if unmarshalErr := json.Unmarshal([]byte(val), &oldConnInfo); unmarshalErr != nil {
			h.Logger.Errorf("unmarshal connectionInfo failed, connID=%s: %v", connID, unmarshalErr)
			// 数据损坏，直接清理
		} else {
			if oldConnInfo.GatewayID != h.gatewayID {
				// 旧连接在其他网关，通过回调触发 MQ 通知
				if h.OnRemoteClose != nil {
					h.OnRemoteClose(oldConnInfo)
				}
			} else {
				// 旧连接在本网关，直接关闭
				h.mu.RLock()
				if oldConn, exists := h.conns[connID]; exists {
					if closeErr := oldConn.Close(); closeErr != nil {
						h.Logger.Errorf("close old local conn failed, connID=%s: %v", connID, closeErr)
					}
				}
				h.mu.RUnlock()
			}
		}
		// 清理旧 Redis 记录
		h.rdb.Del(h.ctx, redisConnKey, redisUserDevKey)
	}

	// 构建新连接信息并写入 Redis（使用 Set 而非 SetNX，因为旧 key 已删除）
	newConnInfo := ConnectionInfo{
		ConnID:    connID,
		UserID:    userID,
		DeviceID:  deviceID,
		GatewayID: h.gatewayID,
		IP:        ip,
		Status:    "online",
	}
	data, err := json.Marshal(newConnInfo)
	if err != nil {
		h.Logger.Errorf("marshal connectionInfo failed, connID=%s: %v", connID, err)
		return nil, errors.WithCode(code.CodeMarshalFailed, "marshal connectionInfo failed")
	}

	pipe := h.rdb.Pipeline()
	pipe.Set(h.ctx, redisConnKey, data, connTTL)
	pipe.Set(h.ctx, redisUserDevKey, connID, connTTL)
	if _, err := pipe.Exec(h.ctx); err != nil {
		h.Logger.Errorf("write connectionInfo to redis failed, connID=%s: %v", connID, err)
		return nil, errors.WithCode(code.CodeCacheSetFailed, "write connectionInfo to redis failed")
	}

	// 继承 Hub 的 context，这样 Hub 关闭时所有连接也会收到信号
	ctx, cancel := context.WithCancel(h.ctx)
	registered := false
	defer func() {
		if !registered {
			cancel()
		}
	}()

	newConn := &Connection{
		ctx:       ctx,
		cancel:    cancel,
		ConnID:    connID,
		UserID:    userID,
		DeviceID:  deviceID,
		GatewayID: h.gatewayID,
		IP:        ip,
		Region:    location,
		conn:      conn,
	}

	h.mu.Lock()
	h.conns[connID] = newConn
	h.mu.Unlock()

	registered = true
	h.Logger.Infof("connection registered: connID=%s gatewayID=%s", connID, h.gatewayID)
	return newConn, nil
}

// Deregister 注销 WebSocket 连接，清理本地和 Redis 记录
// 会先关闭连接再清理记录
func (h *Hub) Deregister(connID string) error {
	// 先从本地 map 删除并关闭连接
	h.mu.Lock()
	if conn, exists := h.conns[connID]; exists {
		_ = conn.Close()
		delete(h.conns, connID)
	}
	h.mu.Unlock()

	// 从 Redis 删除 connectionInfo
	val, err := h.rdb.Get(h.ctx, connKeyPrefix+connID).Result()
	if err == redis.Nil {
		// key 不存在，已清理，忽略
		return nil
	}
	if err != nil {
		h.Logger.Errorf("get connectionInfo from redis failed, connID=%s: %v", connID, err)
		return errors.WithCode(code.CodeCacheGetFailed, "get connectionInfo from redis failed")
	}

	var connInfo ConnectionInfo
	if err := json.Unmarshal([]byte(val), &connInfo); err != nil {
		h.Logger.Errorf("unmarshal connectionInfo failed, connID=%s: %v", connID, err)
		// 数据损坏，强制删除
		h.rdb.Del(h.ctx, connKeyPrefix+connID)
		return errors.WithCode(code.CodeUnmarshalFailed, "unmarshal connectionInfo failed")
	}

	pipe := h.rdb.Pipeline()
	pipe.Del(h.ctx, connKeyPrefix+connID)
	pipe.Del(h.ctx, userDevKey(connInfo.UserID, connInfo.DeviceID))
	if _, err := pipe.Exec(h.ctx); err != nil {
		h.Logger.Errorf("delete connectionInfo from redis failed, connID=%s: %v", connID, err)
		return errors.WithCode(code.CodeCacheDeleteFailed, "delete connectionInfo from redis failed")
	}

	h.Logger.Infof("connection deregistered: connID=%s", connID)
	return nil
}

// GetConn 根据 connID 获取本地连接（不存在返回 nil）
func (h *Hub) GetConn(connID string) *Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.conns[connID]
}

// SendToConn 向指定 connID 的本地连接发送消息
func (h *Hub) SendToConn(connID string, msg []byte) error {
	conn := h.GetConn(connID)
	if conn == nil {
		return errors.WithCode(code.CodeCacheGetFailed, "connection not found: %s", connID)
	}
	return conn.WriteMessage(msg)
}

// WriteMessage 线程安全地向 WebSocket 写入消息
// 如果连接已关闭或 context 已取消，返回错误
func (c *Connection) WriteMessage(msg []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// 检查 context 是否已取消
	select {
	case <-c.ctx.Done():
		return errors.WithCode(code.CodeCacheGetFailed, "connection context canceled")
	default:
	}

	if c.conn == nil {
		return errors.WithCode(code.CodeConnNotFound, "connection already closed")
	}
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}
