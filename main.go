package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" //为了注册postgres驱动
	"github.com/techschool/simplebank/api"
	db "github.com/techschool/simplebank/db/sqlc"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("annnot connect to db :", err)
	}
	// ===== 临时调试代码 =====
	var count int
	err = conn.QueryRow("SELECT COUNT(*) FROM accounts").Scan(&count)
	if err != nil {
		log.Fatal("count query failed:", err)
	}
	log.Println(">>> accounts count =", count)
	// ========================
	store := db.NewStore(conn)

	//创建服务器
	server := api.NewServer(store)
	//启动服务器
	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
