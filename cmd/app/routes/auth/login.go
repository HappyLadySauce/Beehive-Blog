package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// GitHub user info API response (partial fields we need).
// GitHub 用户信息 API 响应中我们需要的部分字段。
type githubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GitHub email API response.
// GitHub 邮箱 API 响应结构。
type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

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
			return nil, fmt.Errorf("invalid username or password")
		}
		return nil, fmt.Errorf("query credential: %w", err)
	}

	// Compare password hash.
	// 比较密码哈希。
	if err := bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Reject disabled or locked accounts.
	// 拒绝被禁用或锁定的账户。
	if user.Status != "active" && user.Status != "pending" {
		return nil, fmt.Errorf("account is %s", user.Status)
	}

	return a.finalizeLogin(&user)
}

// loginByGitHub performs the GitHub OAuth2 authorization code flow.
// loginByGitHub 执行 GitHub OAuth2 授权码流程。
func (a *AuthController) loginByGitHub(ctx *gin.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required for grant_type=%q", v1.GrantTypeGitHubOAuth2)
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

	httpCtx := ctx.Request.Context()
	token, err := oauthCfg.Exchange(httpCtx, req.Code)
	if err != nil {
		klog.ErrorS(err, "Failed to exchange GitHub authorization code")
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	client := oauthCfg.Client(httpCtx, token)

	ghUser, err := fetchGitHubUser(httpCtx, client, cfg.UserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub user info: %w", err)
	}

	email, err := fetchGitHubPrimaryEmail(httpCtx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub email: %w", err)
	}

	user, isNew, err := findOrCreateUser(a.svc.DB, ghUser, email)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user: %w", err)
	}
	if isNew {
		klog.InfoS("Created new user via GitHub OAuth2", "uid", user.ID, "username", user.Username, "email", email)
	}

	return a.finalizeLogin(user)
}

// finalizeLogin updates last_login_at and issues a JWT token pair.
// finalizeLogin 更新 last_login_at 并签发 JWT 令牌对。
func (a *AuthController) finalizeLogin(user *model.User) (*v1.LoginResponse, error) {
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

// fetchGitHubUser calls the GitHub user API and returns the parsed profile.
// fetchGitHubUser 调用 GitHub 用户 API 并返回解析后的用户信息。
func fetchGitHubUser(ctx context.Context, client *http.Client, userInfoURL string) (*githubUser, error) {
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github user API returned status %d", resp.StatusCode)
	}

	var u githubUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("decode github user: %w", err)
	}
	return &u, nil
}

// fetchGitHubPrimaryEmail returns the primary verified email from GitHub.
// fetchGitHubPrimaryEmail 返回 GitHub 的主验证邮箱。
func fetchGitHubPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github emails API returned status %d", resp.StatusCode)
	}

	var emails []githubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("decode github emails: %w", err)
	}

	// Prefer primary+verified, then any verified, then primary.
	// 优先主邮箱+已验证，其次任意已验证，再次主邮箱。
	var fallback string
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
		if e.Verified && fallback == "" {
			fallback = e.Email
		}
		if e.Primary && fallback == "" {
			fallback = e.Email
		}
	}
	if fallback != "" {
		return fallback, nil
	}
	return "", fmt.Errorf("no verified email found on GitHub account; cannot create user")
}

// findOrCreateUser looks up a user by email; if not found, creates one from GitHub profile data.
// findOrCreateUser 按邮箱查找用户；未找到则基于 GitHub 资料创建。
func findOrCreateUser(db *gorm.DB, ghUser *githubUser, email string) (*model.User, bool, error) {
	var user model.User
	err := db.Where("email = ?", email).First(&user).Error
	if err == nil {
		return &user, false, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, false, fmt.Errorf("query user by email: %w", err)
	}

	username := ghUser.Login
	nickname := ghUser.Name
	if nickname == "" {
		nickname = ghUser.Login
	}

	user = model.User{
		Username: username,
		Email:    &email,
		Nickname: &nickname,
		Role:     "member",
		Status:   "active",
	}

	for attempt := 0; attempt < 5; attempt++ {
		result := db.Create(&user)
		if result.Error == nil {
			return &user, true, nil
		}
		if isUniqueViolation(result.Error) && attempt < 4 {
			suffix := rand.Intn(9000) + 1000
			user.Username = fmt.Sprintf("%s_%d", ghUser.Login, suffix)
			continue
		}
		return nil, false, fmt.Errorf("create user: %w", result.Error)
	}

	return nil, false, fmt.Errorf("create user: exceeded retry limit for username %s", ghUser.Login)
}

// isUniqueViolation checks whether the error is a PostgreSQL unique constraint violation (SQLSTATE 23505).
// isUniqueViolation 检查错误是否为 PostgreSQL 唯一约束冲突（SQLSTATE 23505）。
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(msg, "duplicate key")
}
