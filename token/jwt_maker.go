package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32 // 密钥最短长度，防止密钥太短导致安全性不足

// JWTMaker 实现Maker
type JWTMaker struct {
	secretKey string // 用来签名和验证 JWT 的密钥
}

// 给出秘钥，创建一个具体的 JWTMaker 对象，然后把它当作 Maker 接口返回
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize { // 如果密钥长度不够，会返回错误
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey}, nil
}

// CreateToken 根据用户名和有效期，生成一个 JWT token 字符串（用 HS256 算法）
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	// 1. 先创建一个 Payload（包含用户名、签发时间、过期时间等）
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	// 2. 用 HS256 算法和 payload 数据，创建一个还没签名的 JWT token 对象，只是把“头部 + 声明”组装好了
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	// 3. 用密钥对 JWT 进行签名，生成最终的 token 字符串
	return jwtToken.SignedString([]byte(maker.secretKey))
}

// VerifyToken 验证 token 是否有效，如果有效则返回里面的 Payload
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// keyFunc 是一个回调函数，jwt 库在验证时会调用它
	// 作用：确认签名算法是我们期望的 HS256，并返回密钥
	keyFunc := func(token *jwt.Token) (any, error) { //*jwt.Token 就是 jwt 库根据你传入的 token 字符串解析出来的结果。
		// 检查签名算法是否是 HMAC（HS256 属于 HMAC）
		_, ok := token.Method.(*jwt.SigningMethodHMAC) //尝试把它转换成 *jwt.SigningMethodHMAC 类型,如果成功，说明这个 token 用的是 HMAC 算法（允许 HS256 / HS384 / HS512）
		if !ok {
			// 算法不对，说明 token 可能被篡改或使用了错误的算法
			return nil, ErrInvalidToken
		}
		// 返回密钥，供 jwt 库用来验证签名
		return []byte(maker.secretKey), nil
	}

	// 解析并验证 token
	// 第二个参数 &Payload{} 是用来接收解析token出来的数据的
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	/*把 token（Header.Payload.Signature） 拆开，解码 Header 和 Payload。库会把解析出来的东西放到 *jwt.Token类型的数据里 传给你的 keyFunc
	如果 keyFunc 返回了密钥，库就用它去验证签名是否匹配调用 Payload 里的 GetExpirationTime() 等方法，确认 token 是否有效，如果有效就把jwt.Token的值和nil返回给jwtToken和err

	Payload 里的内容就叫 Claims（就是两种说法，本质是同一个东西。但是Claims类型是接口，Payload是结构体，但是实现了这个接口）
	&Payload{}的作用：jwt 库解析 token 后，把里面的 Claims填进这个 Payload 对象，然后库把这个填好的对象放进 jwtToken.Claims，最后返回 jwtToken 给你
	*/
	if err != nil {
		// 如果是过期错误，单独返回过期错误
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		// 其他错误都视为无效 token
		return nil, ErrInvalidToken
	}

	// 把解析出来的 Claims 转换成我们的 Payload 类型（Payload 里的内容就叫 Claims）
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok { //断言失败
		return nil, ErrInvalidToken
	}

	return payload, nil
}
