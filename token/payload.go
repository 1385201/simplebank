package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// VerifyToken函数可能返回的错误：
var (
	ErrExpiredToken = errors.New("token has expired") //token过期
	ErrInvalidToken = errors.New("token is invalid")  //非法token
)

// 这个 Payload 结构体里装token里面需要存放的数据
type Payload struct {
	ID        uuid.UUID `json:"id"`         //用户ID，为了在某些令牌被泄露的情况下使其失效
	Username  string    `json:"username"`   //用户名
	IssuedAt  time.Time `json:"issued_at"`  //签发时间
	ExpiredAt time.Time `json:"expired_at"` //过期时间
}

// 创建一个新的 Payload 对象
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom() //生成一个随机的 UUID（唯一标识符），用作 token 的 ID,比如 550e8400-e29b-41d4-a716-446655440000 这样的字符串
	if err != nil {
		return nil, err
	}
	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}

// Valid 检查 token 是否过期
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}

// 下面这些是 jwt/v5 要求实现的方法
// 返回 token 的过期时间
func (payload *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(payload.ExpiredAt), nil
}

// 返回 token 的签发时间
func (payload *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(payload.IssuedAt), nil
}

// 返回 token 的生效时间（一般不用）
func (payload *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

// 返回签发者（一般不用）
func (payload *Payload) GetIssuer() (string, error) {
	return "", nil
}

// 返回主题/用户标识（一般不用）
func (payload *Payload) GetSubject() (string, error) {
	return "", nil
}

// 返回受众（一般不用）
func (payload *Payload) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}
