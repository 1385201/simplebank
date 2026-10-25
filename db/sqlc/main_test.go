package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries

func TestMain(m *testing.M) {

	// ① 打开数据库连接池,拿到"钥匙"
	//    - dbDriver = "postgres"：告诉 sql 包用哪个驱动
	//    - dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
	//    - sql.Open 只是创建连接池对象,并不真正连库（惰性连接）
	//    - conn 是 *sql.DB,代表到 PostgreSQL 的连接池
	conn, err := sql.Open(dbDriver, dbSource)

	// ② 如果创建连接池失败,打印错误并终止进程
	//    log.Fatal 内部会调用 os.Exit(1),所以这里不用再写 os.Exit
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// ③ 把连接池包装成"数据库操作工具人" testQueries
	//    New(conn) 是 sqlc 生成的构造函数,把 *sql.DB 包装成 *Queries
	//    之后所有测试文件里的 testQueries.CreateAccount(...) 等都通过它调用
	testQueries = New(conn)

	// ④ 开始运行本包所有测试函数,并把结果码交给系统
	//    - m.Run() 依次执行 TestXxx 函数,返回 0（全通过）或 1（有失败）
	//    - os.Exit(...) 必须调用,否则测试结束行为不确定
	os.Exit(m.Run())
}
