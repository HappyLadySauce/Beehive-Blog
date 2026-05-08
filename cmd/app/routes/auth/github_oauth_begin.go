package auth

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/oauth"
)

// GithubOAuthBegin issues a CSRF state, stores it in Redis, and returns the GitHub authorize URL.
// GithubOAuthBegin 签发 CSRF state、写入 Redis，并返回 GitHub 授权 URL。
func (a *AuthController) GithubOAuthBegin(ctx *gin.Context) {
	state, err := oauth.StoreGitHubOAuthState(ctx.Request.Context(), a.svc.Cache, 15*time.Minute)
	if err != nil {
		common.Fail(ctx, common.NewInternal("failed to start oauth session", err))
		return
	}

	cfg := a.svc.Config.GithubOAuth2
	u, err := url.Parse(cfg.AuthURL)
	if err != nil {
		common.Fail(ctx, common.NewInternal("invalid oauth configuration", fmt.Errorf("invalid github auth-url: %w", err)))
		return
	}
	q := u.Query()
	q.Set("client_id", cfg.ClientID)
	q.Set("redirect_uri", cfg.RedirectURL)
	q.Set("scope", "read:user user:email")
	q.Set("state", state)
	u.RawQuery = q.Encode()

	common.Success(ctx, v1.GithubOAuthBeginResponse{
		State:   state,
		AuthURL: u.String(),
	})
}
