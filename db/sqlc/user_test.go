package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/techschool/simplebank/util"
)

func createRandomUser(t *testing.T) User { //是一个测试辅助函数
	hashPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)

	require.NoError(t, err) // 检查：有没有报错？有就失败

	require.NotEmpty(t, user) // 检查：返回的东西是不是空的？

	require.Equal(t, arg.Username, user.Username)             // 检查：名字对不对
	require.Equal(t, arg.HashedPassword, user.HashedPassword) // 检查：密码对不对
	require.Equal(t, arg.FullName, user.FullName)             // 检查：真名对不对
	require.Equal(t, arg.Email, user.Email)                   // 检查：邮箱对不对

	require.True(t, user.PasswordChangeAt.IsZero()) // 检查：上次改密码的时间是不是Go 的 zero time
	require.NotZero(t, user.CreatedAt)              // 检查：有没有生成创建时间
	return user
}

func TestCreateUser(t *testing.T) {

	createRandomUser(t)
}

func TestGetUser(t *testing.T) {

	user1 := createRandomUser(t)

	user2, err := testQueries.GetUser(context.Background(), user1.Username)

	require.NoError(t, err)

	require.NotEmpty(t, user2)

	require.Equal(t, user1.HashedPassword, user2.HashedPassword)

	require.Equal(t, user1.FullName, user2.FullName)

	require.Equal(t, user1.Email, user2.Email)

	require.NotZero(t, user2.CreatedAt)
	require.WithinDuration(t, user1.PasswordChangeAt, user2.PasswordChangeAt, time.Second)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second) //(理论上是一样的，只是数据库的精度和go的精度有点不同，所以用time.second来运行一定的误差)
}
