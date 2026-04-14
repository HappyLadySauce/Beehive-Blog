package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	authutil "github.com/HappyLadySauce/Beehive-Blog/pkg/utils/auth"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/jwt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"k8s.io/klog/v2"
)

const (
	authHeaderName   = "Authorization"
	ctxCurrentUserID = "currentUserId"
	ctxCurrentRole   = "currentUserRole"
	ctxCurrentStatus = "currentUserStatus"
)

// Auth 构建认证中间件：
// 1) 解析并校验 Bearer Token
// 2) 校验 claims 必要字段
// 3) 基于 Redis 进行角色/状态校验
// 4) 将用户信息注入到 gin.Context
func Auth(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svcCtx == nil {
			klog.Error("Auth middleware misconfigured: service context is nil")
			common.AbortFailMessage(c, http.StatusInternalServerError, "auth service unavailable")
			return
		}
		redisClient := svcCtx.Redis
		jwtSecret := svcCtx.Config.JWTOptions.JWTSecret

		if redisClient == nil {
			klog.Error("Auth middleware misconfigured: redis client is nil")
			common.AbortFailMessage(c, http.StatusInternalServerError, "auth service unavailable")
			return
		}
		if strings.TrimSpace(jwtSecret) == "" {
			klog.Error("Auth middleware misconfigured: jwtSecret is empty")
			common.AbortFailMessage(c, http.StatusInternalServerError, "auth service unavailable")
			return
		}

		rawToken, err := authutil.ExtractBearerToken(c.GetHeader(authHeaderName))

		if err != nil {
			common.AbortFailMessage(c, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		claims, err := jwt.ParseToken(jwtSecret, rawToken)
		if err != nil {
			common.AbortFailMessage(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		if claims.UserID <= 0 || strings.TrimSpace(claims.Role) == "" {
			common.AbortFailMessage(c, http.StatusUnauthorized, "invalid token claims")
			return
		}

		authState, ok, err := validateByRedis(c.Request.Context(), redisClient, claims)
		if err != nil {
			klog.ErrorS(err, "Redis auth validation failed", "userID", claims.UserID)
			common.AbortFailMessage(c, http.StatusInternalServerError, "auth service unavailable")
			return
		}
		if !ok {
			common.AbortFailMessage(c, http.StatusUnauthorized, "token is no longer valid")
			return
		}
		c.Set(ctxCurrentUserID, claims.UserID)
		c.Set(ctxCurrentRole, claims.Role)
		c.Set(ctxCurrentStatus, authState.Status)
		c.Next()
	}
}

// RequireRoles 在 Auth 之后使用，用于进行 RBAC 角色拦截。
func RequireRoles(roles ...models.UserRole) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[string(role)] = struct{}{}
	}

	return func(c *gin.Context) {
		role, ok := GetCurrentUserRole(c)
		if !ok {
			common.AbortFailMessage(c, http.StatusUnauthorized, "unauthorized")
			return
		}
		if _, exists := allowed[role]; !exists {
			common.AbortFailMessage(c, http.StatusForbidden, "forbidden")
			return
		}
		c.Next()
	}
}

func GetCurrentUserID(c *gin.Context) (int64, bool) {
	v, exists := c.Get(ctxCurrentUserID)
	if !exists {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

func GetCurrentUserRole(c *gin.Context) (string, bool) {
	v, exists := c.Get(ctxCurrentRole)
	if !exists {
		return "", false
	}
	role, ok := v.(string)
	return role, ok
}

func GetCurrentUserStatus(c *gin.Context) (string, bool) {
	v, exists := c.Get(ctxCurrentStatus)
	if !exists {
		return "", false
	}
	status, ok := v.(string)
	return status, ok
}

type redisAuthState struct {
	Role   string
	Status string
}

func validateByRedis(ctx context.Context, redisClient *redis.Client, claims *jwt.CustomClaims) (redisAuthState, bool, error) {
	key := authutil.UserAuthCacheKey(claims.UserID)
	result, err := redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return redisAuthState{}, false, err
	}
	if len(result) == 0 {
		return redisAuthState{}, false, nil
	}

	role := strings.TrimSpace(result["role"])
	status := strings.TrimSpace(result["status"])
	if role == "" || status == "" {
		return redisAuthState{}, false, nil
	}
	if status != string(models.UserStatusActive) {
		return redisAuthState{}, false, nil
	}
	if role != claims.Role {
		return redisAuthState{}, false, nil
	}
	return redisAuthState{Role: role, Status: status}, true, nil
}

// ValidateBearerToken 执行与 Auth 中间件相同的 JWT + Redis 会话校验，供 WebSocket 等非 HTTP 场景使用。
func ValidateBearerToken(ctx context.Context, svcCtx *svc.ServiceContext, rawToken string) (*jwt.CustomClaims, error) {
	if svcCtx == nil {
		return nil, fmt.Errorf("auth service unavailable")
	}
	redisClient := svcCtx.Redis
	jwtSecret := strings.TrimSpace(svcCtx.Config.JWTOptions.JWTSecret)
	if redisClient == nil || jwtSecret == "" {
		return nil, fmt.Errorf("auth service unavailable")
	}
	claims, err := jwt.ParseToken(jwtSecret, rawToken)
	if err != nil {
		return nil, err
	}
	if claims.UserID <= 0 || strings.TrimSpace(claims.Role) == "" {
		return nil, fmt.Errorf("invalid token claims")
	}
	_, ok, err := validateByRedis(ctx, redisClient, claims)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("token is no longer valid")
	}
	return claims, nil
}
