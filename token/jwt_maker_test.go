package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"github.com/techschool/simplebank/util"
)

// 测试最顺利的情况
func TestJWTMaker(t *testing.T) {
	// 1. 创建 Maker 对象
	maker, err := NewJWTMaker(util.RandomString(32))
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
func TestExpiredJWTToken(t *testing.T) {
	// 1. 创建 Maker 对象
	maker, err := NewJWTMaker(util.RandomString(32))
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
func TestInvalidJWTTokenAlgNone(t *testing.T) {
	// 1. 【制作陷阱】先造一个看起来完全合法的 Payload 数据
	// 比如随便拿一个用户的名字，有效期设为 1 分钟
	payload, err := NewPayload(util.RandomOwner(), time.Minute)
	require.NoError(t, err)

	// 2. 【伪造身份】用黑客的手法生成 Token
	// 注意这里使用的是 jwt.SigningMethodNone 算法（即 alg: none）
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)

	// 3. 【强行签名】
	// 正常的 jwt-go 库为了安全，默认是禁止使用 none 算法的。
	// 只有当你传入这个特定的 unsafe 常量（意思是“我知道很危险，但我偏要用，出了事我负责”），库才会允许生成这种无签名的假 Token。
	// 这就成功伪造出了一个带有合法用户信息的假 Token。
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// 4. 【准备防御】创建一个我们真实的、正规的 JWT Maker（使用正常的 HS256 密钥算法）
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	// 5. 【发起攻击】用我们正规的系统，去验证那个伪造的假 Token
	// 注意：这里用的是 = 而不是 :=，说明是给上面已经声明的 payload 和 err 重新赋值
	payload, err = maker.VerifyToken(token)

	// 6. 【验证防御结果】期待被成功拦截
	// 第一：必须报错（require.Error）
	require.Error(t, err)

	// 第二：报错的具体内容，必须是“无效的 Token”（ErrInvalidToken），而不是其他乱七八糟的错误
	require.EqualError(t, err, ErrInvalidToken.Error())

	// 第三：既然验证失败了，返回的 payload 必须是空的（nil），绝不能把伪造的用户信息放行
	require.Nil(t, payload)

}
