package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// AccessClaims defines JWT claims for access tokens.
type AccessClaims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

// IssueAccessToken builds a signed JWT access token.
func IssueAccessToken(secret, issuer string, userID int64, ttl time.Duration) (string, int64, error) {
	if userID <= 0 {
		return "", 0, errors.New("invalid user id")
	}
	if strings.TrimSpace(secret) == "" {
		return "", 0, errors.New("empty jwt secret")
	}
	now := time.Now()
	expiresAt := now.Add(ttl)
	claims := AccessClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	return signed, expiresAt.Unix(), nil
}

// ParseAccessToken validates and parses a JWT access token.
func ParseAccessToken(secret, tokenString string) (int64, error) {
	if strings.TrimSpace(secret) == "" {
		return 0, errors.New("empty jwt secret")
	}
	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	if !token.Valid || claims.UserID <= 0 {
		return 0, errors.New("invalid token")
	}
	return claims.UserID, nil
}

// BuildRefreshToken generates a cryptographically secure refresh token.
func BuildRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// ExtractBearerToken strips Bearer prefix and returns token payload.
func ExtractBearerToken(header string) (string, error) {
	token := strings.TrimSpace(header)
	if token == "" {
		return "", errors.New("empty authorization header")
	}
	const bearer = "Bearer "
	if strings.HasPrefix(token, bearer) {
		token = strings.TrimSpace(strings.TrimPrefix(token, bearer))
	}
	if token == "" {
		return "", errors.New("empty bearer token")
	}
	return token, nil
}
