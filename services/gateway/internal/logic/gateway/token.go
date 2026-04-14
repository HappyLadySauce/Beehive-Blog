package gateway

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type contextKey string

const AuthHeaderContextKey contextKey = "auth_header"

func parseAccessTokenUserID(token string) (int64, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, fmt.Errorf("empty token")
	}

	const bearerPrefix = "Bearer "
	if strings.HasPrefix(token, bearerPrefix) {
		token = strings.TrimSpace(strings.TrimPrefix(token, bearerPrefix))
	}

	parts := strings.Split(token, ".")
	if len(parts) < 4 || parts[0] != "acc" {
		return 0, fmt.Errorf("invalid token format")
	}

	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || userID <= 0 {
		return 0, fmt.Errorf("invalid token user id")
	}
	return userID, nil
}

func parseAccessTokenUserIDFromContext(ctx context.Context) (int64, error) {
	if ctx == nil {
		return 0, fmt.Errorf("empty context")
	}
	v := ctx.Value(AuthHeaderContextKey)
	header, _ := v.(string)
	return parseAccessTokenUserID(header)
}
