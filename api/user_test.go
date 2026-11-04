package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	mockdb "github.com/techschool/simplebank/db/mock"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/util"
	"go.uber.org/mock/gomock"
)

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)
	testCases := []struct {
		name          string                                                  //测试名称
		body          gin.H                                                   //放的是请求体数据
		buildStubs    func(store *mockdb.MockStore)                           //假的数据库对象的期望操作(必须是数据库层面的错误)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder) //检查函数
	}{
		{
			name: "OK",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) { //输入一个假的数据库对象，然后进行一些列期望操作(必须是数据库层面的错误)
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(user, nil)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username":  "invalid-user#1",
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone) //sql.ErrConnDone是Go 标准库 sql 包里的错误，表示数据库连接断开
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"}) //由于在GO的sql包里没有唯一约束冲突，故用：pq.Error是PostgreSQL 驱动返回的错误类型，Code: "23505"是PostgreSQL 的错误码，23505 专门代表 unique_violation（唯一约束冲突）且*pq.Error 实现了这个错误接口
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     "invalid-email",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TooShortPassword",
			body: gin.H{
				"username":  user.Username,
				"password":  "123",
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) { //在当前这个测试函数里，再单独跑一个小测试，并且给它起个名字
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body) //date现在是[]byte类型了
			require.NoError(t, err)

			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data)) //bytes.NewBuffer(data)将[]byte类型的data变成*bytes.Buffer类型（测试数据（gin.H）→ 序列化成 JSON 字节（[]byte）→ 包装成请求体（io.Reader），才能正确发出去。）
			require.NoError(t, err)                                                      // 要求发送请求时不能报错

			server.router.ServeHTTP(recorder, request) //并把响应记录在 recorder 里

			tc.checkResponse(t, recorder)
		})
	}
}

// 产生一个随机用户，返回给数据库操作用的db.User以及自己传入的密码（如果不返回，则后续找不到，因为db.User里是密文了）
// 测试函数的辅助函数,输入参数为t，
func randomUser(t *testing.T) (user db.User, password string) { //测试时需要明文密码放进请求体（模拟用户提交的表单）哈希后的密码（放在 user.HashedPassword 里）用来设置 mock 的期望，以及检查返回结果
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

// 确保创建成功时，API 返回的用户信息是对的
func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.HashedPassword)
}
