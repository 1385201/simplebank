package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/techschool/simplebank/token"
)

const (
	authorizationHeaderKey  = "authorization"         //请求头的名字。
	authorizationTypeBearer = "bearer"                //支持的认证类型
	authorizationPayloadKey = "authorization_payload" //往 Context 里存用户信息时用的 key
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc { //gin框架的中间件写法
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 { //判断Header里没有authorizationHeader这个字段

			err := errors.New("authorization header is not provided") //构造一个写死字符串的错误
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		fields := strings.Fields(authorizationHeader) //会按空格把字符串切开，返回一个字符串切片
		if len(fields) < 2 {                          //有authorization这个字段但是格式不对（正确的authorization字段的内容格式起码需要有两个字符串中间由空格组成）
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		authorizationType := strings.ToLower(fields[0])   //将认证类型（此处应该是Bearer）统一改成小写，方便比对
		if authorizationType != authorizationTypeBearer { //判断认证类型是不是bearer
			//err := errors.New("invalid authorization header format")
			err := fmt.Errorf("unsupported authorization type %s", authorizationType) //errors.New 只能写死字符串;fmt.Errorf 可以格式化，把变量插进去
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		accessToken := fields[1]                            //能走到这一步说明是bearer类型，那么fields切片里的第二个内容就是token了
		payload, err := tokenMaker.VerifyToken(accessToken) //验证token是否有问题(过期或者不正确)
		if err != nil {                                     //验签
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		ctx.Set(authorizationPayloadKey, payload) //把验证成功后得到的用户信息，临时存到这个请求的上下文里
		ctx.Next()                                //认证通过了，继续执行后面的中间件和真正的业务 Handler
	}
}
