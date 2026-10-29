package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/techschool/simplebank/db/sqlc"
)

type Server struct {
	store  *db.Store   //我用什么处理业务
	router *gin.Engine //我怎么接收和分发请求
}

// NewServer创建一个新的HTTP服务器并设置路由。
func NewServer(store *db.Store) *Server {
	server := &Server{store: store} //只赋值了 store，router 此时是 nil
	router := gin.Default()         //为路由表里的函数提前穿了两层洋葱

	//配置路由表
	router.POST("/accounts", server.createAccount) // POST /accounts    //创建一个账户
	router.GET("/accounts/:id", server.getAccount) // GET /accounts/1	//获取单个account
	router.GET("/accounts", server.listAccount)    // GET /accounts		//获取多个accounts
	router.POST("/accounts/update", server.updateAccount)
	router.GET("/accounts/delete", server.deleteAccount)
	server.router = router
	return server
}

// 启动服务器并监听传入的端口
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// 统一将err返回成json格式
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}

}
