package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// 测试写的password.go有没有问题
func TestPassword(t *testing.T) {
	//密码正确的情况
	password := RandomString(6)
	hashPassword1, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashPassword1)
	err = CheckPassword(password, hashPassword1)
	require.NoError(t, err)

	//密码错误的情况
	wrongPassword := RandomString(6)
	err = CheckPassword(wrongPassword, hashPassword1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	//同一个密码是否会产生不同的哈希
	hashPassword2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashPassword2)
	require.NotEqual(t, hashPassword1, hashPassword2)

}
