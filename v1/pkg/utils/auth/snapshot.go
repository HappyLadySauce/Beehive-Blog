package auth

import (
	"context"
	"errors"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SyncUserAuthSnapshot 同步用户鉴权快照到 Redis。
// - 用户不存在、已软删除、或状态为 deleted 时：删除快照
// - 其他情况：写入 role/status 并刷新 TTL
func SyncUserAuthSnapshot(ctx context.Context, rdb *redis.Client, db *gorm.DB, ttl time.Duration, userID int64) error {
	if rdb == nil || db == nil || userID <= 0 {
		return nil
	}

	authCacheKey := UserAuthCacheKey(userID)

	var user models.User
	err := db.WithContext(ctx).Unscoped().Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return rdb.Del(ctx, authCacheKey).Err()
		}
		return err
	}

	if user.DeletedAt != nil || user.Status == models.UserStatusDeleted {
		return rdb.Del(ctx, authCacheKey).Err()
	}

	if err := rdb.HSet(ctx, authCacheKey, map[string]interface{}{
		"role":   string(user.Role),
		"status": string(user.Status),
	}).Err(); err != nil {
		return err
	}
	return rdb.Expire(ctx, authCacheKey, ttl).Err()
}
