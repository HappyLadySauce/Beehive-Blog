// Package session manages server-side refresh token sessions.
// Package session 管理服务端 refresh token 会话。
package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// ClientMeta records request metadata useful for session audit trails.
// ClientMeta 记录用于会话审计的请求元数据。
type ClientMeta struct {
	IP        string
	UserAgent string
}

// HashRefreshToken returns the SHA-256 hex digest stored in identity.user_sessions.
// HashRefreshToken 返回写入 identity.user_sessions 的 SHA-256 十六进制摘要。
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// NewJTI returns a cryptographically random JWT ID.
// NewJTI 返回密码学安全随机 JWT ID。
func NewJTI() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate refresh jti: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

// IssuePair creates a server-side session and returns access + refresh JWTs bound to it.
// IssuePair 创建服务端会话，并返回绑定该会话的 access + refresh JWT。
func IssuePair(db *gorm.DB, issuer *jwt.Issuer, user *model.User, meta ClientMeta) (jwt.TokenPair, *model.UserSession, error) {
	var pair jwt.TokenPair
	var created model.UserSession
	err := db.Transaction(func(tx *gorm.DB) error {
		p, s, err := IssuePairInTx(tx, issuer, user, meta)
		if err != nil {
			return err
		}
		pair = p
		created = *s
		return nil
	})
	if err != nil {
		return jwt.TokenPair{}, nil, err
	}
	return pair, &created, nil
}

// IssuePairInTx creates a server-side session using an existing transaction.
// IssuePairInTx 使用已有事务创建服务端会话。
func IssuePairInTx(tx *gorm.DB, issuer *jwt.Issuer, user *model.User, meta ClientMeta) (jwt.TokenPair, *model.UserSession, error) {
	if tx == nil {
		return jwt.TokenPair{}, nil, fmt.Errorf("db transaction is nil")
	}
	if issuer == nil {
		return jwt.TokenPair{}, nil, fmt.Errorf("jwt issuer is nil")
	}
	if user == nil || user.ID <= 0 {
		return jwt.TokenPair{}, nil, fmt.Errorf("user is invalid")
	}
	jti, err := NewJTI()
	if err != nil {
		return jwt.TokenPair{}, nil, err
	}

	now := time.Now()
	row := model.UserSession{
		UserID:           user.ID,
		RefreshTokenHash: strings.Repeat("0", 64),
		RefreshJTI:       jti,
		ExpiresAt:        now.Add(issuer.RefreshTTL()),
		CreatedIP:        TruncateString(meta.IP, 64),
		UserAgent:        TruncateString(meta.UserAgent, 512),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := tx.Create(&row).Error; err != nil {
		return jwt.TokenPair{}, nil, fmt.Errorf("create user session: %w", err)
	}

	pair, err := issuer.IssueSessionPair(user.ID, user.Role, row.ID, jti)
	if err != nil {
		return jwt.TokenPair{}, nil, err
	}
	row.RefreshTokenHash = HashRefreshToken(pair.Refresh.Token)
	row.UpdatedAt = time.Now()
	if err := tx.Model(&row).Updates(map[string]any{
		"refresh_token_hash": row.RefreshTokenHash,
		"updated_at":         row.UpdatedAt,
	}).Error; err != nil {
		return jwt.TokenPair{}, nil, fmt.Errorf("store refresh token hash: %w", err)
	}
	return pair, &row, nil
}

// Rotate marks the old session as used and rotated, then issues a replacement session.
// Rotate 标记旧会话已使用并已轮换，然后签发替代会话。
func Rotate(tx *gorm.DB, issuer *jwt.Issuer, oldSession *model.UserSession, user *model.User, meta ClientMeta) (jwt.TokenPair, *model.UserSession, error) {
	now := time.Now()
	if err := tx.Model(oldSession).Updates(map[string]any{
		"last_used_at": &now,
		"rotated_at":   &now,
		"updated_at":   now,
	}).Error; err != nil {
		return jwt.TokenPair{}, nil, fmt.Errorf("rotate old session: %w", err)
	}
	return IssuePairInTx(tx, issuer, user, meta)
}

// RevokeSession marks a session revoked; it is idempotent.
// RevokeSession 标记会话已撤销；该操作具备幂等性。
func RevokeSession(db *gorm.DB, sessionID int64, userID int64, reason string) error {
	if sessionID <= 0 || userID <= 0 {
		return nil
	}
	now := time.Now()
	result := db.Model(&model.UserSession{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", sessionID, userID).
		Updates(map[string]any{
			"revoked_at":     &now,
			"revoked_reason": TruncateString(reason, 64),
			"updated_at":     now,
		})
	if result.Error != nil {
		return fmt.Errorf("revoke session: %w", result.Error)
	}
	return nil
}

// RevokeUserSessions marks all active sessions for a user revoked.
// RevokeUserSessions 标记用户所有活跃会话已撤销。
func RevokeUserSessions(db *gorm.DB, userID int64, reason string) error {
	if userID <= 0 {
		return nil
	}
	now := time.Now()
	result := db.Model(&model.UserSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]any{
			"revoked_at":     &now,
			"revoked_reason": TruncateString(reason, 64),
			"updated_at":     now,
		})
	if result.Error != nil {
		return fmt.Errorf("revoke user sessions: %w", result.Error)
	}
	return nil
}

// TruncateString returns s unchanged if len(s) <= max, otherwise the first max bytes of s.
// TruncateString 在 len(s)<=max 时返回 s 本身，否则返回 s 的前 max 个字节（按字节而非 Unicode 标量）。
// Callers storing into VARCHAR(n) should pass max equal to the column byte budget.
// 写入 VARCHAR(n) 等定宽列时，max 应与列的字节预算一致。
func TruncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
