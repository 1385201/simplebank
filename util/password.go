package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// hashPassword返回密码的bcrypt哈希值（盐+cost+哈希）
func HashPassword(password string) (string, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("fail to hash password: %w", err) //fmt.Errorf("fail to hash password: %w", err)是包装后的错误，本质还是error
	}
	return string(hashPassword), nil
}

// 检查用户输入的密码是否正确
func CheckPassword(password, HashPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(HashPassword), []byte(password))

}
