package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/techschool/simplebank/token"
)

// 在测试中，为 HTTP 请求伪造一个包含有效（或故意无效）Token 的 Authorization 请求头
func addAuthorization(t *testing.T, request *http.Request, tokenMaker token.Maker, authorizationType string, username string, duration time.Duration) {
	token, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token) //构造Authorization请求头
	request.Header.Set(authorizationHeaderKey, authorizationHeader)       //将authorizationHeaderKey:authorizationHeader填入请求头
}

// 测试验证中间件
func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name string //测试名称

		//t的作用是：require.NoError(t, err)之类用到；request作用：必须把请求传进来，才能给它加头；tokenMaker作用：用来生成真正的 Token
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker) //处理
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)           //验证期待
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "unsupported", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			//1.起一个服务器
			server := newTestServer(t, nil) //起一个测试服务器。第二个参数是 db.Store，中间件测试不需要数据库，传 nil

			//2.在起的服务器上添加路由
			authPath := "/auth"
			server.router.GET( // Gin 注册一条带中间件的路由的标准写法
				authPath,
				authMiddleware(server.tokenMaker), // 中间件
				func(ctx *gin.Context) { //handler处理函数
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			//3.构造http请求和响应记录器
			//构造一个假的响应记录器
			recorder := httptest.NewRecorder()

			// 构造一个 原始HTTP 请求：方法 GET，路径 /auth，请求体 nil。
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err) //构造不能出错

			//4.用setupAuth处理（用setupAuth统一封装）
			tc.setupAuth(t, request, server.tokenMaker)

			//5.发出处理后的http请求并交给路由处理（执行请求、走中间件链和 handler）
			server.router.ServeHTTP(recorder, request)

			//6.检查结果是否符合预期（用checkResponse统一封装）
			tc.checkResponse(t, recorder)
		})
	}
}
