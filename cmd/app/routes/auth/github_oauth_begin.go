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
//
// @Summary      Start GitHub OAuth2 (state + authorize URL)
// @Description  Returns a one-time state string and full auth_url for the browser redirect. After GitHub redirects back with code, call POST /api/v1/auth/login with grant_type github_oauth2, code, and the same state.
// @Description  返回一次性 state 与完整 auth_url 供浏览器跳转。GitHub 回调携带 code 后，使用 grant_type=github_oauth2、code 与本 state 调用 POST /api/v1/auth/login。
// @Tags         auth
// @Produce      json
// @Success      200  {object}  common.BaseResponse{data=v1.GithubOAuthBeginResponse}  "state and auth_url"
// @Failure      500  {object}  common.BaseResponse                                  "Redis or configuration error"
// @Router       /api/v1/auth/github/authorize [get]
func (a *AuthController) GithubOAuthBegin(ctx *gin.Context) {
	cfg := a.githubOAuth2Settings()
	if !cfg.Enabled {
		common.Fail(ctx, common.NewForbidden("github oauth2 is disabled", fmt.Errorf("github oauth2 is disabled")))
		return
	}

	state, err := oauth.StoreGitHubOAuthState(ctx.Request.Context(), a.svc.Cache, 15*time.Minute)
	if err != nil {
		common.Fail(ctx, common.NewInternal("failed to start oauth session", err))
		return
	}

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
