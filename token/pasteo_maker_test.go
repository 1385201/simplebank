package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/techschool/simplebank/util"
)

// 测试最顺利的情况
func TestPasetoMaker(t *testing.T) {
	// 1. 创建 Maker 对象
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	// 2. 准备测试数据
	username := util.RandomOwner()
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	// 3. 生成 Token
	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	/*
		为什么这里需要require.NotEmpty而上面对maker却不需要require.NotEmpty？

		说结论对于对象/接口：检查 err 就够了（require.NoError），出错必有 nil
		对于基础数据（如字符串）：即使没有报错，也要检查其内容是否为空（require.NotEmpty），以防逻辑漏洞。
	*/

	//4. 验证 Token
	payload, err := maker.VerifyToken(token)

	// 5.详细比对 Payload 里的各个字段，确保数据解析正确
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

// 测试令牌过期的情况
func TestExpiredPasetoToken(t *testing.T) {
	// 1. 创建 Maker 对象
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	// 2. 准备测试数据
	username := util.RandomOwner()
	duration := -time.Minute
	//issuedAt := time.Now()
	//expiredAt := issuedAt.Add(duration)

	// 3. 生成 Token
	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	//4. 验证 Token
	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error()) //是err.Error()和 ErrExpiredToken.Error()在比
	require.Nil(t, payload)
}

// 测试非法的PasetoToken的情况
func TestInvalidPasetoToken(t *testing.T) {
	// 1. 创建 Maker 对象
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	// 2. 准备测试数据
	username := util.RandomOwner()
	duration := time.Minute

	// 3. 生成 Token
	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	token = ""

	//4. 验证 Token
	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.Nil(t, payload)
}
