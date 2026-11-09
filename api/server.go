package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/token"
	"github.com/techschool/simplebank/util"
)

type Server struct {
	config     util.Config //在创建令牌的时候，需要用到令牌持续时间
	tokenMaker token.Maker //用来生成和验证登录凭证
	store      db.Store    //我用什么处理业务
	router     *gin.Engine //我怎么接收和分发请求

}

// NewServer创建一个新的HTTP服务器并设置路由。
func NewServer(config util.Config, store db.Store) (*Server, error) { //返回一个数据库服务器
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey) //如果想用JWT来签发和验证令牌则把NewPasetoMaker改成NewJWTMaker就行
	if err != nil {
		return nil, fmt.Errorf("can not make token maker:%w", err) //%w表示返回原始的错误
	}
	server := &Server{
		config:     config,
		tokenMaker: tokenMaker,
		store:      store,
	} //只赋值了 store以及tokenMaker，router 此时是 nil,

	//注册新的验证器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency) //你造了一个新安检员（validCurrency），你得带他去跟 Gin 框架的保安队长（binding.Validator.Engine()）报备一下
	}

	//配置路由表
	server.setupRouter()

	return server, nil
}

// 配置路由函数
func (server *Server) setupRouter() {
	router := gin.Default()                       //为路由表里的函数提前穿了两层洋葱
	router.POST("/users", server.createUser)      //创建一个用户
	router.POST("/users/login", server.loginUser) //登录

	router.POST("/accounts", server.createAccount) // POST /accounts    //创建一个账户
	router.GET("/accounts/:id", server.getAccount) // GET /accounts/1	//获取单个account
	router.GET("/accounts", server.listAccount)    // GET /accounts		//获取多个accounts
	router.POST("/accounts/update", server.updateAccount)
	router.GET("/accounts/delete", server.deleteAccount)

	router.POST("/transfers", server.createTransfer)
	server.router = router

}

// 启动服务器并监听传入的端口
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// 统一将err返回成json格式
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}

}
