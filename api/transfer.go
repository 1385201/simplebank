package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/techschool/simplebank/db/sqlc"
)

// 创建转账所需的参数
type transferRequest struct { //从请求体里拿数据，所以设置格式为json
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

// 创建转账（账户1向账户2转账）
func (server *Server) createTransfer(ctx *gin.Context) { //ctx *gin.Context是从 *http.Request 和 http.ResponseWriter 来的，再加一些 Gin 自己的配套数据
	var req transferRequest

	// 1. 把 JSON 绑定到 req 上，并做校验
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) //ctx.JSON是Gin框架返回json响应的方法
		return
	}

	if !server.validAccount(ctx, req.FromAccountID, req.Currency) { //如果账户1的货币类型和req.Currency不一致，返回
		return //跳出 createTransfer 这个处理函数
	}
	if !server.validAccount(ctx, req.ToAccountID, req.Currency) { //如果账户2的货币类型和req.Currency不一致，返回
		return
	}

	// 2. 组装TransferTx参数
	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	// 3. 调用数据库创建transfer事务
	result, err := server.store.TransferTx(ctx, arg) //*gin.Context实现了context.Context接口，所以能传，context.Context接口的作用：Deadline()：获取截止时间  Done()：监听取消信号  Err()：获取取消原因  Value()：读取请求域数据
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// 4. 成功，把result信息返回给客户端
	ctx.JSON(http.StatusOK, result)
}

// 验证accountID账户的货币类型是否和req.Currency保持一致(只有账户1和账户2都和req.Currency保持一致才能进行交易)
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) bool {
	//先根据accountID拿到账户
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil { //要么查询失败，要么查询成功但是没数据
		if err == sql.ErrNoRows { //查询执行成功了，但结果集里一行数据都没有
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false

	}
	//拿到账户但是发现货币种类不一致
	if account.Currency != currency {
		err := fmt.Errorf("account {%d} currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}
	return true
}
