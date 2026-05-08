// Package oauth provides GitHub OAuth2 helpers for user resolution.
// Package oauth 提供 GitHub OAuth2 用户解析相关的辅助函数。
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// GitHubUser is the partial GitHub user API response fields we need.
// GitHubUser 为 GitHub 用户信息 API 响应中我们需要的部分字段。
type GitHubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GitHubEmail is the GitHub email API response.
// GitHubEmail 为 GitHub 邮箱 API 响应结构。
type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// FetchGitHubUser calls the GitHub user API and returns the parsed profile.
// FetchGitHubUser 调用 GitHub 用户 API 并返回解析后的用户信息。
func FetchGitHubUser(ctx context.Context, client *http.Client, userInfoURL string) (*GitHubUser, error) {
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github user API returned status %d", resp.StatusCode)
	}

	var u GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("decode github user: %w", err)
	}
	return &u, nil
}

// FetchGitHubPrimaryEmail returns the primary verified email from GitHub.
// FetchGitHubPrimaryEmail 返回 GitHub 的主验证邮箱。
func FetchGitHubPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github emails API returned status %d", resp.StatusCode)
	}

	var emails []GitHubEmail
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

// FindOrCreateUser looks up a user by email; if not found, creates one from GitHub profile data.
// FindOrCreateUser 按邮箱查找用户；未找到则基于 GitHub 资料创建。
func FindOrCreateUser(db *gorm.DB, ghUser *GitHubUser, email string) (*model.User, bool, error) {
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
