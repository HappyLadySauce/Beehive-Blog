// Package passwd provides bcrypt password hashing and verification.
// Package passwd 提供 bcrypt 密码哈希与验证功能。
package passwd

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// DefaultCost balances security and server load; 12 is current industry default.
// DefaultCost 平衡安全性与服务器负载；12 为当前业界默认值。
const DefaultCost = 12

// Hash returns a bcrypt hash of the plaintext password using DefaultCost.
// Hash 使用 DefaultCost 对明文密码进行 bcrypt 哈希。
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(bytes), nil
}

// Verify compares a plaintext password against a bcrypt hash.
// Verify 将明文密码与 bcrypt 哈希进行比对。
func Verify(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
