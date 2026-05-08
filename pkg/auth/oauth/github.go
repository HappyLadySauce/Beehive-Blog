// Package oauth provides GitHub OAuth2 helpers for user resolution.
// Package oauth 提供 GitHub OAuth2 用户解析相关的辅助函数。
package oauth

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

const githubProvider = "github"

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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
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
	const emailsURL = "https://api.github.com/user/emails"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, emailsURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
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
	if ghUser == nil || ghUser.ID <= 0 {
		return nil, false, fmt.Errorf("github user id is required")
	}
	subject := strconv.FormatInt(ghUser.ID, 10)

	boundUser, found, err := findUserByProviderSubject(db, githubProvider, subject)
	if err != nil {
		return nil, false, err
	}
	if found {
		return boundUser, false, nil
	}

	var user model.User
	err = db.Where("email = ?", email).First(&user).Error
	if err == nil {
		if err := bindProviderIdentity(db, user.ID, githubProvider, subject, email); err != nil {
			return nil, false, err
		}
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
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			return bindProviderIdentity(tx, user.ID, githubProvider, subject, email)
		})
		if err == nil {
			return &user, true, nil
		}
		if isUniqueViolation(err) && attempt < 4 {
			suffix, err := randomUsernameSuffix()
			if err != nil {
				return nil, false, fmt.Errorf("create user: %w", err)
			}
			user.Username = fmt.Sprintf("%s_%d", ghUser.Login, suffix)
			continue
		}
		if user, found, findErr := findUserByProviderSubject(db, githubProvider, subject); findErr != nil {
			return nil, false, findErr
		} else if found {
			return user, false, nil
		}
		return nil, false, fmt.Errorf("create user: %w", err)
	}

	return nil, false, fmt.Errorf("create user: exceeded retry limit for username %s", ghUser.Login)
}

func findUserByProviderSubject(db *gorm.DB, provider, subject string) (*model.User, bool, error) {
	var identity model.UserIdentity
	err := db.Where("provider = ? AND provider_subject = ?", provider, subject).First(&identity).Error
	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("query user identity: %w", err)
	}
	var user model.User
	if err := db.First(&user, identity.UserID).Error; err != nil {
		return nil, false, fmt.Errorf("query identity user: %w", err)
	}
	return &user, true, nil
}

func bindProviderIdentity(db *gorm.DB, userID int64, provider, subject, email string) error {
	now := time.Now()
	identity := model.UserIdentity{
		UserID:          userID,
		Provider:        provider,
		ProviderSubject: subject,
		EmailAtBind:     &email,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&identity).Error; err != nil {
		if isUniqueViolation(err) {
			return nil
		}
		return fmt.Errorf("bind user identity: %w", err)
	}
	return nil
}

// isUniqueViolation checks whether the error is a PostgreSQL unique constraint violation (SQLSTATE 23505).
// isUniqueViolation 检查错误是否为 PostgreSQL 唯一约束冲突（SQLSTATE 23505）。
// randomUsernameSuffix returns a value in [1000, 9999] using crypto/rand.
// randomUsernameSuffix 使用 crypto/rand 返回 [1000, 9999] 内的整数。
func randomUsernameSuffix() (int, error) {
	var buf [2]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return 0, err
	}
	n := binary.BigEndian.Uint16(buf[:])
	return int(n%9000) + 1000, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(msg, "duplicate key")
}
