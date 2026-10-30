package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" //为了注册postgres驱动
	"github.com/techschool/simplebank/api"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/util"
)

func main() {
	config, err := util.LoadConfig(".") //从当前文件夹，加载配置环境
	if err != nil {
		log.Fatal("can not  load configurations", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("annnot connect to db :", err)
	}

	store := db.NewStore(conn)

	//创建服务器
	server := api.NewServer(store)
	//启动服务器
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
