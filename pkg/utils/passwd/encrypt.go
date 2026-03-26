package passwd

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost bcrypt 加密成本 (10-14 之间，越高越安全但越慢)
	BcryptCost = 12
)

// HashPassword 使用 bcrypt 加密密码
// bcrypt 内部会自动生成随机盐值，无需外部传入
func HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}

	// bcrypt 会自动生成盐值并包含在哈希结果中
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword 验证密码是否正确
func VerifyPassword(password, hashedPassword string) bool {
	if len(password) == 0 || len(hashedPassword) == 0 {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// IsHashedPassword 检查字符串是否已经是 bcrypt 哈希格式
func IsHashedPassword(str string) bool {
	// bcrypt 哈希以 $2a$, $2b$, $2x$, $2y$ 开头
	return len(str) == 60 && (str[:4] == "$2a$" || str[:4] == "$2b$" || str[:4] == "$2x$" || str[:4] == "$2y$")
}
