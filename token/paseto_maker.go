package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

// PasetoMaker 实现了通用的 Maker 接口，专门用于生成和验证 PASETO 格式的 Token
type PasetoMaker struct {

	// 这里指定为 *paseto.V2，代表我们使用的是 PASETO 协议的第 2 版本。
	// (PASETO V2 底层使用的是 XChaCha20-Poly1305 对称加密算法，既保证了机密性又保证了完整性)
	paseto *paseto.V2

	// 因为我们的 Token 是在本地使用,所以用对称加密的密钥（比如单体服务或内部微服务之间的鉴权），
	// 不需要让第三方去验证，所以使用对称密钥：加密和验证都用这同一把钥匙。
	// (注：PASETO V2 要求这个密钥的长度必须严格等于 32 字节)
	symmetricKey []byte
}

func NewPasetoMaker(symmetricKey string) (Maker, error) {
	// 校验密钥长度：PASETO V2 要求必须严格等于 32 字节
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	// 初始化 PasetoMaker 实例
	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),       // 实例化 V2 加密引擎
		symmetricKey: []byte(symmetricKey), // 字符串转为字节切片
	}
	return maker, nil
}

// CreateToken 生成 PASETO Token
func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	// 1. 创建包含用户信息的 Payload
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	// 2. 调用底层 V2 引擎，使用对称密钥将 payload 加密成字符串
	// 第三个参数 nil 表示不需要附加 Footer（页脚数据）
	return maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
}

// VerifyToken 验证并解析 PASETO Token
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	// 1. 创建一个空的 Payload 容器，用来接收解密后的数据
	payload := &Payload{}

	// 2. 调用底层 V2 引擎进行解密，将解出的数据填入 payload 中
	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		// 解密失败（如密钥不对、Token被篡改），直接返回错误
		return nil, err
	}

	// 3. 检查 Payload 是否已经过期
	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	// 4. 一切正常，返回有效的数据
	return payload, nil
}
