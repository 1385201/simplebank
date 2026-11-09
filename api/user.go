package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/util"
)

// 客户端发过来的 JSON 长什么样的定义
type createUserRequest struct { //从请求体里拿数据，所以设置格式为json
	Username string `json:"username" binding:"required,alphanum"` //alphanum是标签（build-in tags）保证符合ascii编码
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"` //email是标签（build-in tags）,保证邮箱的格式
}

type userResponse struct { //因为之前的db.User里有hash_password字段，容易造成安全损失
	Username string `json:"username"`

	FullName         string    `json:"full_name"`
	Email            string    `json:"email"`
	PasswordChangeAt time.Time `json:"password_change_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// 根据数据库里得到的user来构建响应给客户端的userResponse结构体
func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:         user.Username,
		FullName:         user.FullName,
		Email:            user.Email,
		PasswordChangeAt: user.PasswordChangeAt,
		CreatedAt:        user.CreatedAt,
	}
}

// 接收 JSON → 校验 → 转成数据库参数 → 写库 → 返回结果。
func (server *Server) createUser(ctx *gin.Context) { //ctx *gin.Context是从 *http.Request 和 http.ResponseWriter 来的，再加一些 Gin 自己的配套数据
	var req createUserRequest

	// 1. 把 JSON 绑定到 req 上，并做校验
	if err := ctx.ShouldBindJSON(&req); err != nil {

		ctx.JSON(http.StatusBadRequest, errorResponse(err)) //ctx.JSON是Gin框架返回json响应的方法
		return
	}
	hashPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// 2. 组装数据库建立user操作所需要参数
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}
	//arg = db.CreateUserParams{}
	// 3. 调用数据库创建用户
	user, err := server.store.CreateUser(ctx, arg) //*gin.Context实现了context.Context接口，所以能传，context.Context接口的作用：Deadline()：获取截止时间  Done()：监听取消信号  Err()：获取取消原因  Value()：读取请求域数据
	//有可能是数据库内部的问题也有可能是违反了数据库设置的唯一约束或者外键约束
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok { //通过类型断言，把错误转换成*pq.Err，如果转换成功则返回ture，否则false，换言之：只有当这个错误确实是 PostgreSQL 返回的错误时，才会进入 if 里面
			switch pqErr.Code.Name() {
			case "unique_violation": //唯一约束冲突（email和username是唯一的）
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rep := newUserResponse(user)
	// 4. 成功，把账户信息返回给客户端
	ctx.JSON(http.StatusOK, rep)
}

type loginUserRequest struct { //从请求体里拿数据，装到这个结构体里，所以设置格式为json
	Username string `json:"username" binding:"required,alphanum"` //alphanum是标签（build-in tags）保证符合ascii编码
	Password string `json:"password" binding:"required,min=6"`
}
type loginUserResponse struct { //返回给客户端的结构体
	AccessToken string       `json:"access_token"`
	User        userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) //404表示资源不存在（也可以写401）
		return
	}
	user, err := server.store.GetUser(ctx, req.Username) //从数据库查询该用户
	if err != nil {
		if err == sql.ErrNoRows { //不存在该用户
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		//服务器内部错误
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	//检查密码是否正确
	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err)) //不写BadRequest，是因为请求本身是合规的，只是我不认识你（身份验证失败）
		return
	}
	accessToken, err := server.tokenMaker.CreateToken(req.Username, server.config.AccessTokenDuration)
	if err != nil { //创建token失败
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rep := loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, rep)

}
