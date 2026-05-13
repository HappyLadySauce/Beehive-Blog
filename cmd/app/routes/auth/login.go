package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/oauth"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/passwd"
	authsession "github.com/HappyLadySauce/Beehive-Blog/pkg/auth/session"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// Login handles POST /api/v1/auth/login.
// Login 处理 POST /api/v1/auth/login。
//
// @Summary      Login (local password or GitHub OAuth2)
// @Description  Authenticates by grant_type. Use grant_type=local with account (username or email) and password, or grant_type=github_oauth2 with code and state from GET /api/v1/auth/github/authorize. 中文：按 grant_type 登录；local 使用用户名或邮箱与密码；github_oauth2 使用授权码与 state。
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      v1.LoginRequest  true  "Login request"
// @Success      200   {object}  common.BaseResponse{data=v1.LoginResponse}  "Issued access and refresh tokens"
// @Failure      400   {object}  common.BaseResponse                         "Validation error or unsupported grant_type"
// @Failure      401   {object}  common.BaseResponse                         "Invalid credentials or invalid OAuth code/state"
// @Failure      403   {object}  common.BaseResponse                         "Account status disallows login"
// @Failure      500   {object}  common.BaseResponse                         "Internal error"
// @Router       /api/v1/auth/login [post]
func (a *AuthController) Login(ctx *gin.Context) {
	var req v1.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}

	var (
		resp *v1.LoginResponse
		err  error
	)
	switch req.GrantType {
	case v1.GrantTypeLocal:
		resp, err = a.loginByLocal(ctx.Request.Context(), &req, clientMetaFromGin(ctx))
	case v1.GrantTypeGitHubOAuth2:
		resp, err = a.loginByGitHub(ctx.Request.Context(), &req, clientMetaFromGin(ctx))
	default:
		err = common.NewBadRequest("unsupported grant_type", fmt.Errorf("unsupported grant_type: %q", req.GrantType))
	}

	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// loginByLocal authenticates a local user via account (username or email) and password.
// loginByLocal 通过账户（用户名或邮箱）和密码认证本地用户。
func (a *AuthController) loginByLocal(ctx context.Context, req *v1.LoginRequest, meta authsession.ClientMeta) (*v1.LoginResponse, error) {
	if req.Account == "" || req.Password == "" {
		return nil, common.NewBadRequest("account and password are required", nil)
	}

	// Look up user by username or email among live rows.
	// 在活跃行中按用户名或邮箱查找。
	var user model.User
	if err := a.svc.DB.Where("username = ? OR email = ?", req.Account, req.Account).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NewUnauthorized("invalid credentials", nil)
		}
		return nil, common.NewInternal("failed to login", fmt.Errorf("query user: %w", err))
	}

	// Look up active credential; return identical error to avoid user enumeration.
	// 查找活跃凭证；返回相同错误以防范用户枚举攻击。
	var cred model.UserCredential
	if err := a.svc.DB.Where("user_id = ?", user.ID).First(&cred).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NewUnauthorized("invalid credentials", nil)
		}
		return nil, common.NewInternal("failed to login", fmt.Errorf("query credential: %w", err))
	}

	// Compare password hash.
	// 比较密码哈希。
	if err := passwd.Verify(req.Password, cred.PasswordHash); err != nil {
		return nil, common.NewUnauthorized("invalid credentials", nil)
	}

	return a.finalizeLogin(ctx, &user, meta)
}

// loginByGitHub performs the GitHub OAuth2 authorization code flow.
// loginByGitHub 执行 GitHub OAuth2 授权码流程。
func (a *AuthController) loginByGitHub(ctx context.Context, req *v1.LoginRequest, meta authsession.ClientMeta) (*v1.LoginResponse, error) {
	if req.Code == "" {
		return nil, common.NewBadRequest("code is required", nil)
	}

	if req.State == "" {
		return nil, common.NewUnauthorized("invalid or expired oauth session", nil)
	}
	ok, err := oauth.ConsumeGitHubOAuthState(ctx, a.svc.Cache, req.State)
	if err != nil {
		klog.ErrorS(err, "Failed to consume GitHub OAuth state")
		return nil, common.NewUnauthorized("invalid or expired oauth session", err)
	}
	if !ok {
		return nil, common.NewUnauthorized("invalid or expired oauth session", nil)
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

	token, err := oauthCfg.Exchange(ctx, req.Code)
	if err != nil {
		klog.ErrorS(err, "Failed to exchange GitHub authorization code")
		return nil, common.NewUnauthorized("failed to exchange authorization code", err)
	}

	client := oauthCfg.Client(ctx, token)

	ghUser, err := oauth.FetchGitHubUser(ctx, client, cfg.UserInfoURL)
	if err != nil {
		return nil, common.NewUnauthorized("failed to fetch GitHub user info", err)
	}

	email, err := oauth.FetchGitHubPrimaryEmail(ctx, client)
	if err != nil {
		return nil, common.NewUnauthorized("failed to fetch GitHub email", err)
	}

	user, isNew, err := oauth.FindOrCreateUser(a.svc.DB, ghUser, email)
	if err != nil {
		return nil, common.NewInternal("failed to resolve oauth user", err)
	}
	if isNew {
		klog.InfoS("Created new user via GitHub OAuth2", "uid", user.ID, "username", user.Username, "email", email)
	}

	return a.finalizeLogin(ctx, user, meta)
}

// assertUserMayLogin rejects non-loginable account statuses for every auth path.
// assertUserMayLogin 在所有认证路径上拒绝不可登录的账户状态。
func assertUserMayLogin(user *model.User) error {
	if user.Status != "active" && user.Status != "pending" {
		return common.NewForbidden("account is not allowed to login", fmt.Errorf("account is %s", user.Status))
	}
	return nil
}

// finalizeLogin updates last_login_at and issues a JWT token pair.
// finalizeLogin 更新 last_login_at 并签发 JWT 令牌对。
func (a *AuthController) finalizeLogin(ctx context.Context, user *model.User, meta authsession.ClientMeta) (*v1.LoginResponse, error) {
	if err := assertUserMayLogin(user); err != nil {
		return nil, err
	}
	now := time.Now()
	if err := a.svc.DB.Model(user).Update("last_login_at", now).Error; err != nil {
		klog.ErrorS(err, "Failed to update last_login_at", "uid", user.ID)
	}

	pair, _, err := authsession.IssuePair(a.svc.DB, a.svc.Token, user, meta)
	if err != nil {
		return nil, common.NewInternal("failed to issue token", err)
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

func clientMetaFromGin(ctx *gin.Context) authsession.ClientMeta {
	return authsession.ClientMeta{
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	}
}
