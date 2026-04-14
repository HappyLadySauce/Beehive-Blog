package svc

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"
)

type userRecord struct {
	ID           int64
	Username     string
	Nickname     string
	Email        string
	Role         string
	PasswordHash string
}

type memoryStore struct {
	mu sync.RWMutex

	nextUserID int64
	usersByID  map[int64]*userRecord
	accountIdx map[string]int64
	refreshIdx map[string]int64
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		nextUserID: 1,
		usersByID:  make(map[int64]*userRecord),
		accountIdx: make(map[string]int64),
		refreshIdx: make(map[string]int64),
	}
}

func (s *memoryStore) Register(username, nickname, email, password string) (*userRecord, string, string, int64, error) {
	keyUsername := accountKey(username)
	keyEmail := accountKey(email)
	if keyUsername == "" || keyEmail == "" || strings.TrimSpace(password) == "" {
		return nil, "", "", 0, fmt.Errorf("invalid register request")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.accountIdx[keyUsername]; ok {
		return nil, "", "", 0, fmt.Errorf("username already exists")
	}
	if _, ok := s.accountIdx[keyEmail]; ok {
		return nil, "", "", 0, fmt.Errorf("email already exists")
	}

	id := s.nextUserID
	s.nextUserID++
	if nickname == "" {
		nickname = username
	}

	user := &userRecord{
		ID:           id,
		Username:     username,
		Nickname:     nickname,
		Email:        email,
		Role:         "owner",
		PasswordHash: passwordDigest(password),
	}

	s.usersByID[id] = user
	s.accountIdx[keyUsername] = id
	s.accountIdx[keyEmail] = id

	accessToken, expiresIn, err := buildAccessToken(id)
	if err != nil {
		return nil, "", "", 0, err
	}
	refreshToken, err := buildRefreshToken(id)
	if err != nil {
		return nil, "", "", 0, err
	}
	s.refreshIdx[refreshToken] = id

	return copyUser(user), accessToken, refreshToken, expiresIn, nil
}

func (s *memoryStore) Login(account, password string) (*userRecord, string, string, int64, error) {
	key := accountKey(account)
	if key == "" || strings.TrimSpace(password) == "" {
		return nil, "", "", 0, fmt.Errorf("invalid login request")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	uid, ok := s.accountIdx[key]
	if !ok {
		return nil, "", "", 0, fmt.Errorf("account not found")
	}
	user := s.usersByID[uid]
	if user == nil || user.PasswordHash != passwordDigest(password) {
		return nil, "", "", 0, fmt.Errorf("invalid credentials")
	}

	accessToken, expiresIn, err := buildAccessToken(user.ID)
	if err != nil {
		return nil, "", "", 0, err
	}
	refreshToken, err := buildRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", 0, err
	}
	s.refreshIdx[refreshToken] = user.ID

	return copyUser(user), accessToken, refreshToken, expiresIn, nil
}

func (s *memoryStore) Refresh(refreshToken string) (*userRecord, string, string, int64, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, "", "", 0, fmt.Errorf("empty refresh token")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	userID, ok := s.refreshIdx[refreshToken]
	if !ok {
		return nil, "", "", 0, fmt.Errorf("refresh token invalid")
	}
	user := s.usersByID[userID]
	if user == nil {
		return nil, "", "", 0, fmt.Errorf("user not found")
	}
	delete(s.refreshIdx, refreshToken)

	accessToken, expiresIn, err := buildAccessToken(user.ID)
	if err != nil {
		return nil, "", "", 0, err
	}
	newRefreshToken, err := buildRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", 0, err
	}
	s.refreshIdx[newRefreshToken] = user.ID

	return copyUser(user), accessToken, newRefreshToken, expiresIn, nil
}

func (s *memoryStore) GetUser(userID int64) (*userRecord, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.usersByID[userID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return copyUser(user), nil
}

func accountKey(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func passwordDigest(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func buildAccessToken(userID int64) (string, int64, error) {
	return buildToken("acc", userID, 7200)
}

func buildRefreshToken(userID int64) (string, error) {
	token, _, err := buildToken("ref", userID, 86400*30)
	return token, err
}

func buildToken(prefix string, userID int64, ttlSeconds int64) (string, int64, error) {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return "", 0, err
	}
	expiresIn := time.Now().Unix() + ttlSeconds
	token := fmt.Sprintf("%s.%d.%d.%s", prefix, userID, expiresIn, base64.RawURLEncoding.EncodeToString(buf))
	return token, expiresIn, nil
}

func copyUser(in *userRecord) *userRecord {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}
