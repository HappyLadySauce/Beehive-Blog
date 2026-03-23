// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package gateway

import (
	"net/http"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/ip"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/ws"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 写超时时间
	writeWait = 10 * time.Second
	// Pong 超时时间
	pongWait = 60 * time.Second
	// Ping 发送间隔（必须小于 pongWait）
	pingPeriod = (pongWait * 9) / 10
	// 最大消息大小
	maxMessageSize = 4096
)

func WsEntryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// 从 query 参数获取 userID/deviceID，生产环境应替换为鉴权结果
		userID := r.URL.Query().Get("userID")
		deviceID := r.URL.Query().Get("deviceID")
		if userID == "" || deviceID == "" {
			http.Error(w, "missing userID/deviceID", http.StatusBadRequest)
			return
		}

		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("websocket upgrade failed: %v", err)
			return
		}

		// 获取客户端 IP
		clientIP := ip.GetClientIP(r)

		// 查询 IP 地理位置
		location := ""
		if svcCtx.IPManager != nil {
			loc := svcCtx.IPManager.MustQuery(clientIP)
			location = loc.String()
			logx.Infof("IP %s location: %s", clientIP, location)
		}

		conn, err := svcCtx.Hub.Register(wsConn, userID, deviceID, clientIP, location)
		if err != nil {
			logx.Errorf("register connection failed: userID=%s deviceID=%s err=%v", userID, deviceID, err)
			wsConn.Close()
			return
		}

		// 确保退出时注销连接（Deregister 会调用 conn.Close()）
		defer func() {
			if err := svcCtx.Hub.Deregister(conn.ConnID); err != nil {
				logx.Errorf("deregister connection failed: connID=%s err=%v", conn.ConnID, err)
			}
		}()

		logx.Infof("connection established: connID=%s", conn.ConnID)

		// 启动心跳检测（读协程）
		go readPump(conn)

		// TODO: 消息处理逻辑
		// 可以在此处添加业务消息处理，或者启动单独的协程处理
	}
}

// readPump 读取客户端消息并处理心跳
// 监听 WebSocket 的读取，处理 Ping/Pong，检测连接是否存活
func readPump(conn *ws.Connection) {
	defer func() {
		_ = conn.Close()
	}()

	wsConn := conn.RawConn()
	if wsConn == nil {
		return
	}

	// 设置读取限制和超时
	wsConn.SetReadLimit(maxMessageSize)
	_ = wsConn.SetReadDeadline(time.Now().Add(pongWait))

	// 设置 Pong 处理器
	wsConn.SetPongHandler(func(string) error {
		_ = wsConn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// 监听连接 context 取消
	ctx := conn.Context()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			messageType, data, err := wsConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logx.Errorf("read error: connID=%s err=%v", conn.ConnID, err)
				}
				return
			}

			// TODO: 处理业务消息
			logx.Infof("received message: connID=%s type=%d len=%d", conn.ConnID, messageType, len(data))
		}
	}
}
