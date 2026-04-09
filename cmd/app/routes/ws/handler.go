package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middlewares"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/archives"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	authutil "github.com/HappyLadySauce/Beehive-Blog/pkg/utils/auth"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"
)

const (
	maxWSConnectionsPerUser  = 5
	maxAutosaveMsgsPerMinute = 120
	readWait                 = 120 * time.Second
	writeWait                = 15 * time.Second
	maxMessageBytes          = 1 << 20
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return checkOrigin(r.Header.Get("Origin"))
		},
	}

	globalHub       = newHub(maxWSConnectionsPerUser)
	autosaveLimiter = newSlidingWindowLimiter(maxAutosaveMsgsPerMinute, time.Minute)
)

// clientMsg 浏览器下行 JSON。
type clientMsg struct {
	Type      string          `json:"type"`
	RequestID string          `json:"requestId,omitempty"`
	ArticleID int64           `json:"articleId"`
	Payload   json.RawMessage `json:"payload"`
}

// serverMsg 服务端上行 JSON。
type serverMsg struct {
	Type      string          `json:"type"`
	RequestID string          `json:"requestId,omitempty"`
	Code      int             `json:"code,omitempty"`
	Message   string          `json:"message,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// Init 在根路由注册 GET /api/v1/ws，不经过 v1 全局限流中间件。
func Init(svcCtx *svc.ServiceContext) {
	articleSvc := archives.NewArticleAdmin(svcCtx)
	r := router.Router()
	r.GET("/api/v1/ws", handleWebSocket(svcCtx, articleSvc))
}

func extractToken(c *gin.Context) string {
	if t := strings.TrimSpace(c.Query("token")); t != "" {
		return t
	}
	if h := c.GetHeader("Authorization"); h != "" {
		if raw, err := authutil.ExtractBearerToken(h); err == nil {
			return raw
		}
	}
	return ""
}

func handleWebSocket(svcCtx *svc.ServiceContext, articleSvc *archives.ArticleAdmin) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			common.FailMessage(c, http.StatusUnauthorized, "missing token")
			return
		}
		claims, err := middlewares.ValidateBearerToken(c.Request.Context(), svcCtx, token)
		if err != nil {
			common.FailMessage(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if claims.Role != string(models.UserRoleAdmin) {
			common.FailMessage(c, http.StatusForbidden, "forbidden")
			return
		}
		if !globalHub.tryAdd(claims.UserID) {
			common.FailMessage(c, http.StatusTooManyRequests, "too many websocket connections for this user")
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			globalHub.remove(claims.UserID)
			klog.ErrorS(err, "websocket upgrade failed", "userID", claims.UserID)
			return
		}

		conn.SetReadLimit(maxMessageBytes)
		if err := conn.SetReadDeadline(time.Now().Add(readWait)); err != nil {
			klog.ErrorS(err, "websocket SetReadDeadline")
		}
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(readWait))
		})

		go serveConn(conn, claims.UserID, articleSvc)
	}
}

func serveConn(conn *websocket.Conn, userID int64, articleSvc *archives.ArticleAdmin) {
	defer func() {
		globalHub.remove(userID)
		_ = conn.Close()
	}()

	now := time.Now()
	if now.Unix()%120 == 0 {
		autosaveLimiter.cleanup(now)
	}

	for {
		if err := conn.SetReadDeadline(time.Now().Add(readWait)); err != nil {
			return
		}
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				klog.V(4).InfoS("websocket read closed", "userID", userID, "err", err)
			}
			return
		}

		var msg clientMsg
		if err := json.Unmarshal(data, &msg); err != nil {
			writeJSON(conn, serverMsg{Type: "error", Code: http.StatusBadRequest, Message: "invalid json"})
			continue
		}

		switch strings.TrimSpace(msg.Type) {
		case "ping":
			writeJSON(conn, serverMsg{Type: "pong"})
		case "article.autosave":
			handleArticleAutosave(conn, userID, articleSvc, &msg)
		default:
			writeJSON(conn, serverMsg{Type: "error", Code: http.StatusBadRequest, Message: "unknown message type"})
		}
	}
}

func handleArticleAutosave(conn *websocket.Conn, userID int64, articleSvc *archives.ArticleAdmin, msg *clientMsg) {
	key := "uid:" + strconv.FormatInt(userID, 10)
	if !autosaveLimiter.allow(key, time.Now()) {
		writeJSON(conn, serverMsg{
			Type:      "article.autosave.result",
			RequestID: msg.RequestID,
			Code:      http.StatusTooManyRequests,
			Message:   "too many autosave messages, please slow down",
		})
		return
	}

	if msg.ArticleID <= 0 {
		writeJSON(conn, serverMsg{
			Type:      "article.autosave.result",
			RequestID: msg.RequestID,
			Code:      http.StatusBadRequest,
			Message:   "invalid article id",
		})
		return
	}

	var req v1.UpdateArticleRequest
	if len(msg.Payload) > 0 && string(msg.Payload) != "null" {
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			writeJSON(conn, serverMsg{
				Type:      "article.autosave.result",
				RequestID: msg.RequestID,
				Code:      http.StatusBadRequest,
				Message:   "invalid payload",
			})
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, code, err := articleSvc.UpdateArticle(ctx, msg.ArticleID, &req, nil)
	if err != nil {
		writeJSON(conn, serverMsg{
			Type:      "article.autosave.result",
			RequestID: msg.RequestID,
			Code:      code,
			Message:   err.Error(),
		})
		return
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		klog.ErrorS(err, "ws: marshal article detail")
		writeJSON(conn, serverMsg{
			Type:      "article.autosave.result",
			RequestID: msg.RequestID,
			Code:      http.StatusInternalServerError,
			Message:   "internal server error",
		})
		return
	}

	writeJSON(conn, serverMsg{
		Type:      "article.autosave.result",
		RequestID: msg.RequestID,
		Code:      http.StatusOK,
		Data:      raw,
	})
}

func writeJSON(conn *websocket.Conn, v serverMsg) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	return conn.WriteJSON(v)
}
