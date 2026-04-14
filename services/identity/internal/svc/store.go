package svc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/libs/security"
	"github.com/HappyLadySauce/Beehive-Blog/services/identity/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type userRecord struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	Nickname     string `db:"nickname"`
	Email        string `db:"email"`
	Role         string `db:"role"`
	PasswordHash string `db:"password_hash"`
}

type identityStore struct {
	conn       sqlx.SqlConn
	redis      *redis.Redis
	authConfig config.AuthConf
}

func newIdentityStore(conn sqlx.SqlConn, redisClient *redis.Redis, authCfg config.AuthConf) (*identityStore, error) {
	s := &identityStore{
		conn:       conn,
		redis:      redisClient,
		authConfig: authCfg,
	}
	if err := s.ensureSchema(context.Background()); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *identityStore) Register(ctx context.Context, username, nickname, email, password string) (*userRecord, string, string, int64, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if username == "" || email == "" || password == "" {
		return nil, "", "", 0, fmt.Errorf("invalid register request")
	}
	if len(password) < 8 {
		return nil, "", "", 0, fmt.Errorf("password too short")
	}
	if nickname = strings.TrimSpace(nickname); nickname == "" {
		nickname = username
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, "", "", 0, err
	}

	var inserted userRecord
	query := `
INSERT INTO users (username, nickname, email, role, password_hash)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, username, nickname, email, role, password_hash`
	if err = s.conn.QueryRowCtx(ctx, &inserted, query, username, nickname, email, "owner", passwordHash); err != nil {
		if isUniqueViolation(err) {
			return nil, "", "", 0, fmt.Errorf("username or email already exists")
		}
		return nil, "", "", 0, err
	}

	accessToken, expiresIn, err := security.IssueAccessToken(s.authConfig.AccessSecret, s.authConfig.Issuer, inserted.ID, s.authConfig.AccessTTL)
	if err != nil {
		return nil, "", "", 0, err
	}
	refreshToken, err := security.BuildRefreshToken()
	if err != nil {
		return nil, "", "", 0, err
	}
	if err = s.storeRefreshToken(ctx, refreshToken, inserted.ID); err != nil {
		return nil, "", "", 0, err
	}
	return &inserted, accessToken, refreshToken, expiresIn, nil
}

func (s *identityStore) Login(ctx context.Context, account, password string) (*userRecord, string, string, int64, error) {
	account = strings.TrimSpace(account)
	password = strings.TrimSpace(password)
	if account == "" || password == "" {
		return nil, "", "", 0, fmt.Errorf("invalid login request")
	}

	var user userRecord
	query := `
SELECT id, username, nickname, email, role, password_hash
FROM users
WHERE lower(username) = lower($1) OR lower(email) = lower($1)
LIMIT 1`
	if err := s.conn.QueryRowCtx(ctx, &user, query, account); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, "", "", 0, fmt.Errorf("account not found")
		}
		return nil, "", "", 0, err
	}
	if err := verifyPassword(user.PasswordHash, password); err != nil {
		return nil, "", "", 0, fmt.Errorf("invalid credentials")
	}

	accessToken, expiresIn, err := security.IssueAccessToken(s.authConfig.AccessSecret, s.authConfig.Issuer, user.ID, s.authConfig.AccessTTL)
	if err != nil {
		return nil, "", "", 0, err
	}
	refreshToken, err := security.BuildRefreshToken()
	if err != nil {
		return nil, "", "", 0, err
	}
	if err = s.storeRefreshToken(ctx, refreshToken, user.ID); err != nil {
		return nil, "", "", 0, err
	}
	return &user, accessToken, refreshToken, expiresIn, nil
}

func (s *identityStore) Refresh(ctx context.Context, refreshToken string) (*userRecord, string, string, int64, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, "", "", 0, fmt.Errorf("empty refresh token")
	}

	rawUserID, err := s.redis.GetDelCtx(ctx, refreshTokenKey(refreshToken))
	if err != nil || rawUserID == "" {
		return nil, "", "", 0, fmt.Errorf("refresh token invalid")
	}
	userID, err := strconv.ParseInt(rawUserID, 10, 64)
	if err != nil || userID <= 0 {
		return nil, "", "", 0, fmt.Errorf("refresh token invalid")
	}

	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, "", "", 0, err
	}
	accessToken, expiresIn, err := security.IssueAccessToken(s.authConfig.AccessSecret, s.authConfig.Issuer, user.ID, s.authConfig.AccessTTL)
	if err != nil {
		return nil, "", "", 0, err
	}
	newRefreshToken, err := security.BuildRefreshToken()
	if err != nil {
		return nil, "", "", 0, err
	}
	if err = s.storeRefreshToken(ctx, newRefreshToken, user.ID); err != nil {
		return nil, "", "", 0, err
	}
	return user, accessToken, newRefreshToken, expiresIn, nil
}

func (s *identityStore) GetUser(ctx context.Context, userID int64) (*userRecord, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	var user userRecord
	query := `
SELECT id, username, nickname, email, role, password_hash
FROM users
WHERE id = $1
LIMIT 1`
	if err := s.conn.QueryRowCtx(ctx, &user, query, userID); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (s *identityStore) ensureSchema(ctx context.Context) error {
	const query = `
CREATE TABLE IF NOT EXISTS users (
	id BIGSERIAL PRIMARY KEY,
	username VARCHAR(64) NOT NULL UNIQUE,
	nickname VARCHAR(128) NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	role VARCHAR(32) NOT NULL DEFAULT 'owner',
	password_hash VARCHAR(255) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`
	_, err := s.conn.ExecCtx(ctx, query)
	return err
}

func (s *identityStore) storeRefreshToken(ctx context.Context, token string, userID int64) error {
	ttl := int(s.authConfig.RefreshTTL.Seconds())
	if ttl <= 0 {
		ttl = int((30 * 24 * time.Hour).Seconds())
	}
	return s.redis.SetexCtx(ctx, refreshTokenKey(token), strconv.FormatInt(userID, 10), ttl)
}

func hashPassword(raw string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func verifyPassword(hashed, raw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(raw))
}

func refreshTokenKey(token string) string {
	return "identity:refresh:" + token
}

func isUniqueViolation(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
