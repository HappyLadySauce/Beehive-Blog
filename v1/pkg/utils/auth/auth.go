package auth

import (
	"errors"
	"fmt"
	"strings"
)

const bearerPrefix = "Bearer "

// ExtractBearerToken parses Authorization header and returns bearer token.
func ExtractBearerToken(authHeader string) (string, error) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("missing bearer token")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
	if token == "" {
		return "", errors.New("empty token")
	}

	return token, nil
}

// UserAuthCacheKey returns redis key for user auth snapshot.
func UserAuthCacheKey(userID int64) string {
	return fmt.Sprintf("auth:user:%d", userID)
}
