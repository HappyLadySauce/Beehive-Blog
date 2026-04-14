package gateway

import (
	"context"
	"fmt"
)

type contextKey string

const AuthUserIDContextKey contextKey = "auth_user_id"

func parseAccessTokenUserIDFromContext(ctx context.Context) (int64, error) {
	if ctx == nil {
		return 0, fmt.Errorf("empty context")
	}
	v := ctx.Value(AuthUserIDContextKey)
	userID, ok := v.(int64)
	if !ok || userID <= 0 {
		return 0, fmt.Errorf("missing user id in context")
	}
	return userID, nil
}
