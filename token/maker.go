package token

import "time"

// Maker是管理tokens（令牌）的一个接口
type Maker interface { //这样可以在不同类型的token makers之间进行切换（有点类似真假数据库Store）

	//根据用户名和有效期生成 token
	CreateToken(username string, duration time.Duration) (string, error)

	//验证 token，成功就返回里面的数据（Payload）
	VerifyToken(token string) (*Payload, error)
}
