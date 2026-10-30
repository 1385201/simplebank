package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	mockdb "github.com/techschool/simplebank/db/mock" // 引入刚才用 mockgen 生成的假数据库包
	db "github.com/techschool/simplebank/db/sqlc"     // 引入真实的数据库模型（Account 结构体等）
	"github.com/techschool/simplebank/util"           // 引入工具包（生成随机数据）
	"go.uber.org/mock/gomock"                         // 引入 gomock 测试框架
)

// TestGetAccountAPI 测试 GET /accounts/:id 这个 API
func TestGetAccountAPI(t *testing.T) {
	// 1. 准备测试数据：生成一个随机的、假的账户对象
	account := randomAccount()

	// 2. 初始化 Mock 控制器
	// gomock.NewController(t) 负责管理整个测试过程中所有 Mock 的预期行为
	ctrl := gomock.NewController(t)
	// defer ctrl.Finish() 非常重要！它会在测试结束前检查：你期待调用的函数，到底有没有被调用？
	defer ctrl.Finish()

	// 创建一个假的数据库实例（MockStore），它实现了和真实数据库一样的接口
	store := mockdb.NewMockStore(ctrl)

	// 3. 设定预期：告诉假数据库，等下遇到什么情况该怎么做
	store.EXPECT(). // 期待调用 GetAccount 方法
			GetAccount(gomock.Any(), gomock.Eq(account.ID)). // 参数1：任意 Context；参数2：必须等于我们生成的 account.ID
			Times(1).                                        // 约束次数，在整个测试过程中，必须被调用，且只能被调用 1 次（导演ctrl在盯着）
			Return(account, nil)                             // 约束结果，当被调用时，直接返回我们刚才造的 account 对象，且没有错误(nil)，

	// 4. 启动测试服务器并发起 HTTP 请求
	// 把假数据库塞进服务器里！因为 NewServer 接收的是 Store 接口，所以它能完美接纳 MockStore
	server := NewServer(store)
	// httptest.NewRecorder() 相当于一个“录音机”，用来记录 API 返回的响应（状态码、响应体等）
	recorder := httptest.NewRecorder()

	// 构造请求的 URL，比如 "/accounts/717"
	url := fmt.Sprintf("/accounts/%d", account.ID)
	// 创建一个 GET 请求，因为 GET 请求没有请求体，所以第三个参数是 nil
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err) // 要求发送请求时不能报错

	// 真正让 Gin 路由器函数ServeHTTP处理这个请求，并把响应记录在 recorder 里（不需要真的启动一个网络端口）
	server.router.ServeHTTP(recorder, request)

	// 5. 验证结果
	// 验证返回的 HTTP 状态码是不是 200 (StatusOK)
	require.Equal(t, http.StatusOK, recorder.Code)
	// 验证返回的响应体（JSON）里的数据，是不是和我们一开始造的 account 一模一样
	requireBodyMatchAccount(t, recorder.Body, account)
}

// randomAccount 随机生成一个账户对象，专门供测试使用
func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000), // 随机 ID：1 到 1000 之间
		Owner:    util.RandomOwner(),      // 随机拥有者名字
		Balance:  util.RandomMoney(),      // 随机余额
		Currency: util.RandomCurrency(),   // 随机货币单位 (USD/EUR)
	}
}

// requireBodyMatchAccount 用于比对：响应体里的账户，和我们预期的是否一致
func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	// 从“录音机”里把所有返回的 JSON 数据读取出来
	// （注：ioutil.ReadAll 在新版 Go 中推荐写成 io.ReadAll，不过写法逻辑是一样的）
	date, err := io.ReadAll(body)
	require.NoError(t, err) // 读取过程不能报错

	// 准备一个空的变量，用来装反序列化后的账户数据
	var gotAccount db.Account
	// 把读取出来的 JSON 字符串（date），解析成 Go 语言里的结构体对象（gotAccount）
	err = json.Unmarshal(date, &gotAccount)
	require.NoError(t, err) // 解析过程不能报错

	// 终极比对：期望的 account 结构体，和实际拿到的 gotAccount 结构体，必须完全相等！
	require.Equal(t, account, gotAccount)
}
