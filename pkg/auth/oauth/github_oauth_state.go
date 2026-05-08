// GitHub OAuth2 CSRF state storage and single-use consumption in Redis.
// GitHub OAuth2 CSRF state 在 Redis 中的存储与一次性消费。
package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// githubOAuthStateKeyPrefix namespaces OAuth state keys in Redis.
// githubOAuthStateKeyPrefix 为 Redis 中 OAuth state 键的前缀命名空间。
const githubOAuthStateKeyPrefix = "beehive:oauth:github:state:"

// StoreGitHubOAuthState generates a cryptographically random state and stores it in Redis with TTL.
// StoreGitHubOAuthState 生成密码学安全的 state 并以 TTL 写入 Redis。
func StoreGitHubOAuthState(ctx context.Context, rdb *redis.Client, ttl time.Duration) (state string, err error) {
	if rdb == nil {
		return "", fmt.Errorf("redis client is nil")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("ttl must be > 0")
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate state: %w", err)
	}
	state = hex.EncodeToString(raw)
	key := githubOAuthStateKeyPrefix + state
	if err := rdb.Set(ctx, key, "1", ttl).Err(); err != nil {
		return "", fmt.Errorf("store oauth state: %w", err)
	}
	return state, nil
}

// ConsumeGitHubOAuthState removes one matching state key and reports whether it existed (single-use).
// ConsumeGitHubOAuthState 删除匹配的 state 键并返回其是否存在（一次性消费）。
func ConsumeGitHubOAuthState(ctx context.Context, rdb *redis.Client, state string) (consumed bool, err error) {
	if rdb == nil {
		return false, fmt.Errorf("redis client is nil")
	}
	if state == "" {
		return false, nil
	}
	key := githubOAuthStateKeyPrefix + state
	val, err := rdb.GetDel(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("consume oauth state: %w", err)
	}
	return val != "", nil
}
