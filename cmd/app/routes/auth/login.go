package auth

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/oauth"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/passwd"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// Login dispatches based on grant_type to the appropriate authentication method.
// Login 根据 grant_type 分发到对应的认证方法。
func (a *AuthController) Login(ctx *gin.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	switch req.GrantType {
	case v1.GrantTypeLocal:
		return a.loginByLocal(ctx, req)
	case v1.GrantTypeGitHubOAuth2:
		return a.loginByGitHub(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported grant_type: %q", req.GrantType)
	}
}

// loginByLocal authenticates a local user via account (username or email) and password.
// loginByLocal 通过账户（用户名或邮箱）和密码认证本地用户。
func (a *AuthController) loginByLocal(ctx *gin.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	if req.Account == "" || req.Password == "" {
		return nil, fmt.Errorf("account and password are required for grant_type=%q", v1.GrantTypeLocal)
	}

	// Look up user by username or email among live rows.
	// 在活跃行中按用户名或邮箱查找。
	var user model.User
	if err := a.svc.DB.Where("username = ? OR email = ?", req.Account, req.Account).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid account or password")
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	// Look up active credential; return identical error to avoid user enumeration.
	// 查找活跃凭证；返回相同错误以防范用户枚举攻击。
	var cred model.UserCredential
	if err := a.svc.DB.Where("user_id = ?", user.ID).First(&cred).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid account or password")
		}
		return nil, fmt.Errorf("query credential: %w", err)
	}

	// Compare password hash.
	// 比较密码哈希。
	if err := passwd.Verify(req.Password, cred.PasswordHash); err != nil {
		return nil, fmt.Errorf("invalid account or password")
	}

	return a.finalizeLogin(&user)
}

// loginByGitHub performs the GitHub OAuth2 authorization code flow.
// loginByGitHub 执行 GitHub OAuth2 授权码流程。
func (a *AuthController) loginByGitHub(ctx *gin.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required for grant_type=%q", v1.GrantTypeGitHubOAuth2)
	}

	httpCtx := ctx.Request.Context()

	if req.State == "" {
		return nil, fmt.Errorf("invalid or expired oauth session")
	}
	ok, err := oauth.ConsumeGitHubOAuthState(httpCtx, a.svc.Cache, req.State)
	if err != nil {
		klog.ErrorS(err, "Failed to consume GitHub OAuth state")
		return nil, fmt.Errorf("invalid or expired oauth session")
	}
	if !ok {
		return nil, fmt.Errorf("invalid or expired oauth session")
	}

	cfg := a.svc.Config.GithubOAuth2
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
	}

	token, err := oauthCfg.Exchange(httpCtx, req.Code)
	if err != nil {
		klog.ErrorS(err, "Failed to exchange GitHub authorization code")
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	client := oauthCfg.Client(httpCtx, token)

	ghUser, err := oauth.FetchGitHubUser(httpCtx, client, cfg.UserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub user info: %w", err)
	}

	email, err := oauth.FetchGitHubPrimaryEmail(httpCtx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub email: %w", err)
	}

	user, isNew, err := oauth.FindOrCreateUser(a.svc.DB, ghUser, email)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user: %w", err)
	}
	if isNew {
		klog.InfoS("Created new user via GitHub OAuth2", "uid", user.ID, "username", user.Username, "email", email)
	}

	return a.finalizeLogin(user)
}

// assertUserMayLogin rejects non-loginable account statuses for every auth path.
// assertUserMayLogin 在所有认证路径上拒绝不可登录的账户状态。
func assertUserMayLogin(user *model.User) error {
	if user.Status != "active" && user.Status != "pending" {
		return fmt.Errorf("account is %s", user.Status)
	}
	return nil
}

// finalizeLogin updates last_login_at and issues a JWT token pair.
// finalizeLogin 更新 last_login_at 并签发 JWT 令牌对。
func (a *AuthController) finalizeLogin(user *model.User) (*v1.LoginResponse, error) {
	if err := assertUserMayLogin(user); err != nil {
		return nil, err
	}
	now := time.Now()
	if err := a.svc.DB.Model(user).Update("last_login_at", now).Error; err != nil {
		klog.ErrorS(err, "Failed to update last_login_at", "uid", user.ID)
	}

	pair, err := a.svc.Token.IssuePair(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to issue JWT: %w", err)
	}

	return &v1.LoginResponse{
		Token: v1.AuthToken{
			AccessToken:  pair.Access.Token,
			TokenType:    pair.TokenType,
			ExpiresIn:    pair.Access.ExpiresIn,
			RefreshToken: pair.Refresh.Token,
		},
	}, nil
}
