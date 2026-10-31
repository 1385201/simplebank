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
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) //ctx.JSON是Gin框架返回json响应的方法
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
	ID int64 `uri:"id" binding:"required,min=1"` //从URL 路径里拿数据，所以用uri:"id"，并限制id的最小值是1

}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil { //如果ctx里的数据不满足getAccountRequest结构体的限制，则会报错
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID) //如果server.store是SQLStore，那就会去真数据库查询，如果是MockStore，那就会去假数据库查询（真假数据库都实现了Querier接口，都有GetAccount函数）
	if err != nil {                                      //要么查询失败，要么查询成功但是没数据
		if err == sql.ErrNoRows { //查询执行成功了，但结果集里一行数据都没有
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return

	}
	//account = db.Account{}

	ctx.JSON(http.StatusOK, account)

}

type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"` //从查询参数里获取数据用form
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.ListAccountsParams{
		Limit:  req.PageSize,                    //一页的账户数目
		Offset: (req.PageID - 1) * req.PageSize, //偏移量
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return

	}

	ctx.JSON(http.StatusOK, accounts)

}

type updateAccountRequest struct {
	ID      int64 `json:"id" binding:"required,min=1"`
	Balance int64 `json:"balance" binding:"required"`
}

func (server *Server) updateAccount(ctx *gin.Context) {
	var req updateAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.UpdateAccountParams{
		ID:      req.ID,
		Balance: req.Balance,
	}

	account, err := server.store.UpdateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return

	}

	ctx.JSON(http.StatusOK, account)

}

type deleteAccount struct {
	ID int64 `form:"id" binding:"required,min=1"`
}

func (server *Server) deleteAccount(ctx *gin.Context) {
	var req deleteAccount
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}

	err := server.store.DeleteAccount(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, "delete success!")
}
