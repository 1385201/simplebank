package api

/*在 Go 语言中，如果你在测试文件里定义了一个 func TestMain(m *testing.M)那么当你运行这个包的测试（比如点击 run package tests）时，
Go 不会直接开始跑测试用例，而是先跑 TestMain。你可以在里面做各种“全局初始化”工作（比如连测试数据库、加载环境变量）默认情况下 Gin 是 DebugMode，
切换到 TestMode 后，Gin 变得非常安静，只在必要的时候输出*/
import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/util"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// 辅助测试函数,第一个参数都是t
func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}
	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}
