package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/techschool/simplebank/db/sqlc"
)

// 客户端发过来的 JSON 长什么样的定义
type createAccountRequest struct { //从请求体里拿数据，所以设置格式为json
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR "`
}

// 接收 JSON → 校验 → 转成数据库参数 → 写库 → 返回结果。
func (server *Server) createAccount(ctx *gin.Context) { //ctx *gin.Context是从 *http.Request 和 http.ResponseWriter 来的，再加一些 Gin 自己的配套数据
	var req createAccountRequest

	// 1. 把 JSON 绑定到 req 上，并做校验
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// 2. 组装数据库参数
	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0, // 新账户余额固定为 0
	}

	// 3. 调用数据库创建账户
	account, err := server.store.CreateAccount(ctx, arg) //*gin.Context实现了context.Context接口，所以能传，context.Context接口的作用：Deadline()：获取截止时间  Done()：监听取消信号  Err()：获取取消原因  Value()：读取请求域数据
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// 4. 成功，把账户信息返回给客户端
	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"` //从URL 路径里拿数据，所以用uri:"id"

}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil { //要么查询失败，要么查询成功但是没数据
		if err == sql.ErrNoRows { //查询执行成功了，但结果集里一行数据都没有
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return

	}

	ctx.JSON(http.StatusOK, account)

}
