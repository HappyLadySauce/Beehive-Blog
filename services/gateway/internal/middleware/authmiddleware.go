package middleware

import (
	"context"
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/libs/security"
	gwlogic "github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/logic/gateway"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
)

type AuthMiddleware struct {
	svcCtx *svc.ServiceContext
}

func NewAuthMiddleware(svcCtx *svc.ServiceContext) *AuthMiddleware {
	return &AuthMiddleware{svcCtx: svcCtx}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := security.ExtractBearerToken(r.Header.Get("Authorization"))
		if err != nil {
			writeJSONError(r.Context(), w, http.StatusUnauthorized, "AUTH_REQUIRED", "authorization is required")
			return
		}

		userID, err := security.ParseAccessToken(m.svcCtx.Config.Auth.AccessSecret, token)
		if err != nil {
			writeJSONError(r.Context(), w, http.StatusUnauthorized, "AUTH_INVALID", "invalid access token")
			return
		}

		ctx := context.WithValue(r.Context(), gwlogic.AuthUserIDContextKey, userID)
		next(w, r.WithContext(ctx))
	}
}
