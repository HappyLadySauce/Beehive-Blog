// Package jwt issues and verifies HMAC-SHA256 JWT credentials.
// Package jwt 提供 HMAC-SHA256 JWT 凭证的签发与验证能力。
package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	jwt5 "github.com/golang-jwt/jwt/v5"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

// TokenTypeBearer is the OAuth2 bearer scheme advertised in responses.
// TokenTypeBearer 为响应中声明的 OAuth2 Bearer 模式。
const TokenTypeBearer = "Bearer"

// Token-use values isolate access and refresh tokens to prevent confusion attacks.
// Token-use 取值用于隔离访问与刷新令牌，避免混用攻击。
const (
	TokenUseAccess  = "access"
	TokenUseRefresh = "refresh"
)

// ErrWrongTokenUse indicates a token was used in a context expecting a different "use" claim.
// ErrWrongTokenUse 表示令牌的 use 声明与当前调用场景不匹配。
var ErrWrongTokenUse = errors.New("token use mismatch")

// Claims is the JWT payload signed by Issuer; UID and Role are app-specific, Use isolates token classes.
// Claims 为 Issuer 签发的 JWT 负载；UID/Role 为应用自定义字段，Use 隔离令牌类别。
type Claims struct {
	UID  int64  `json:"uid"`
	Role string `json:"role"`
	Use  string `json:"use"`
	jwt5.RegisteredClaims
}

// SignedToken is an issued JWT plus its lifetime expressed in seconds.
// SignedToken 表示一份签发后的 JWT 及其有效期（秒）。
type SignedToken struct {
	Token     string
	ExpiresIn int64
}

// TokenPair groups an access token with its refresh token and the bearer scheme.
// TokenPair 将访问令牌、刷新令牌与 Bearer 方案打包返回。
type TokenPair struct {
	Access    SignedToken
	Refresh   SignedToken
	TokenType string
}

// Issuer signs and verifies HS256 JWTs; safe for concurrent use after construction.
// Issuer 负责 HS256 JWT 的签发与验证；构造完成后并发安全。
type Issuer struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewIssuer wires an Issuer from JWTOptions; callers must pass already-validated options.
// NewIssuer 基于已通过 Validate 的 JWTOptions 构造 Issuer。
func NewIssuer(opts *options.JWTOptions) (*Issuer, error) {
	if opts == nil {
		return nil, fmt.Errorf("jwt options is nil")
	}
	if opts.Issuer == "" {
		return nil, fmt.Errorf("jwt issuer is empty")
	}
	if len(opts.Secret) == 0 {
		return nil, fmt.Errorf("jwt secret is empty")
	}
	if opts.AccessTTL <= 0 || opts.RefreshTTL <= 0 {
		return nil, fmt.Errorf("jwt ttls must be > 0")
	}
	return &Issuer{
		secret:     []byte(opts.Secret),
		issuer:     opts.Issuer,
		accessTTL:  opts.AccessTTL,
		refreshTTL: opts.RefreshTTL,
	}, nil
}

// AccessTTL returns the configured access token lifetime; useful for response shaping.
// AccessTTL 返回访问令牌的存活时长，便于在响应层使用。
func (i *Issuer) AccessTTL() time.Duration { return i.accessTTL }

// RefreshTTL returns the configured refresh token lifetime.
// RefreshTTL 返回刷新令牌的存活时长。
func (i *Issuer) RefreshTTL() time.Duration { return i.refreshTTL }

// IssuePair signs an access + refresh token pair for the given subject and role.
// IssuePair 为指定主体与角色签发 access + refresh 令牌对。
func (i *Issuer) IssuePair(uid int64, role string) (TokenPair, error) {
	if uid <= 0 {
		return TokenPair{}, fmt.Errorf("uid must be > 0, got %d", uid)
	}
	access, err := i.sign(uid, role, TokenUseAccess, i.accessTTL)
	if err != nil {
		return TokenPair{}, fmt.Errorf("sign access token: %w", err)
	}
	refresh, err := i.sign(uid, role, TokenUseRefresh, i.refreshTTL)
	if err != nil {
		return TokenPair{}, fmt.Errorf("sign refresh token: %w", err)
	}
	return TokenPair{
		Access:    SignedToken{Token: access, ExpiresIn: int64(i.accessTTL / time.Second)},
		Refresh:   SignedToken{Token: refresh, ExpiresIn: int64(i.refreshTTL / time.Second)},
		TokenType: TokenTypeBearer,
	}, nil
}

// IssueAccess signs only an access token; useful for refresh flows that rotate access only.
// IssueAccess 仅签发访问令牌，适用于只轮换 access 的刷新流程。
func (i *Issuer) IssueAccess(uid int64, role string) (SignedToken, error) {
	if uid <= 0 {
		return SignedToken{}, fmt.Errorf("uid must be > 0, got %d", uid)
	}
	tok, err := i.sign(uid, role, TokenUseAccess, i.accessTTL)
	if err != nil {
		return SignedToken{}, fmt.Errorf("sign access token: %w", err)
	}
	return SignedToken{Token: tok, ExpiresIn: int64(i.accessTTL / time.Second)}, nil
}

// Parse verifies an HS256 token, asserts iss + alg, and returns the decoded Claims.
// Parse 验证 HS256 令牌、校验 iss 与签名算法，并返回解码后的 Claims。
func (i *Issuer) Parse(tokenString string) (*Claims, error) {
	parsed, err := jwt5.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt5.Token) (any, error) {
			if _, ok := t.Method.(*jwt5.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method %q", t.Header["alg"])
			}
			return i.secret, nil
		},
		jwt5.WithIssuer(i.issuer),
		jwt5.WithValidMethods([]string{jwt5.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// ParseAccess parses a token and rejects anything that is not an access token.
// ParseAccess 解析令牌并拒绝任何非 access 令牌。
func (i *Issuer) ParseAccess(tokenString string) (*Claims, error) {
	claims, err := i.Parse(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Use != TokenUseAccess {
		return nil, fmt.Errorf("%w: want %q, got %q", ErrWrongTokenUse, TokenUseAccess, claims.Use)
	}
	return claims, nil
}

// ParseRefresh parses a token and rejects anything that is not a refresh token.
// ParseRefresh 解析令牌并拒绝任何非 refresh 令牌。
func (i *Issuer) ParseRefresh(tokenString string) (*Claims, error) {
	claims, err := i.Parse(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Use != TokenUseRefresh {
		return nil, fmt.Errorf("%w: want %q, got %q", ErrWrongTokenUse, TokenUseRefresh, claims.Use)
	}
	return claims, nil
}

// sign builds and signs a single JWT with the given use claim and TTL.
// sign 构造并签发一份带有给定 use 声明与 TTL 的 JWT。
func (i *Issuer) sign(uid int64, role, use string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UID:  uid,
		Role: role,
		Use:  use,
		RegisteredClaims: jwt5.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   strconv.FormatInt(uid, 10),
			IssuedAt:  jwt5.NewNumericDate(now),
			NotBefore: jwt5.NewNumericDate(now),
			ExpiresAt: jwt5.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt5.NewWithClaims(jwt5.SigningMethodHS256, claims).SignedString(i.secret)
}

// Context keys set by AuthMiddleware for downstream handlers.
// AuthMiddleware 为下游处理器注入的 Context 键。
const (
	// ClaimsKey stores the parsed *Claims.
	// ClaimsKey 存储解析后的 *Claims。
	ClaimsKey = "claims"
	// UIDKey stores the authenticated user ID as int64.
	// UIDKey 存储已认证用户 ID（int64）。
	UIDKey = "uid"
	// RoleKey stores the authenticated user role as string.
	// RoleKey 存储已认证用户角色（string）。
	RoleKey = "role"
)

// GetClaims extracts the *Claims from the Gin context; returns nil if absent.
// GetClaims 从 Gin 上下文提取 *Claims；不存在时返回 nil。
func GetClaims(ctx *gin.Context) *Claims {
	v, ok := ctx.Get(ClaimsKey)
	if !ok {
		return nil
	}
	claims, ok := v.(*Claims)
	if !ok {
		return nil
	}
	return claims
}
