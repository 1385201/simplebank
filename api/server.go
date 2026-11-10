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
	//上面的部分不需要进行任何类型的授权，而下面的部分需要

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker)) //该路由组下的路由都会自动通过认证中间件authMiddleware的检查

	authRoutes.POST("/accounts", server.createAccount) // POST /accounts    //创建一个账户
	authRoutes.GET("/accounts/:id", server.getAccount) // GET /accounts/1	//获取单个account
	authRoutes.GET("/accounts", server.listAccount)    // GET /accounts		//获取多个accounts
	authRoutes.POST("/accounts/update", server.updateAccount)
	authRoutes.GET("/accounts/delete", server.deleteAccount)

	authRoutes.POST("/transfers", server.createTransfer) //保存的是全部路包括由用 router 直接注册的（公开接口：创建用户、登录）以及用 authRoutes 注册的（需要认证的接口：账户、转账等）
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
