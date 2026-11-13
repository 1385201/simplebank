postgres:# 创建一个名为 postgres14 的 PostgreSQL 容器，端口 5432，用户 root，密码 secret
	docker run --name postgres14 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:14-alpine

createdb:# 在 postgres14 容器中创建名为 simple_bank 的数据库
	docker exec -it postgres14 createdb --username=root --owner=root simple_bank

dropdb:# 删除 postgres14 容器中的 simple_bank 数据库
	docker exec -it postgres14 dropdb simple_bank

migrateup:# 执行数据库迁移，把 db/migration 下的 SQL 应用到数据库（建表等）
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up
migrateup1:# 执行数据库迁移，更新当前版本的下一个版本（注意脏数据）
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:# 回滚数据库迁移，撤销 migrateup 做的变更
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down
migratedown1:# 回滚数据库迁移，回到当前版本的上一个版本（注意脏数据）
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

sqlc:# 根据 SQL 文件生成 Go 代码（db/sqlc 下的文件）
	sqlc generate

test:# 运行所有测试，并显示覆盖率
	go test -v -cover ./...
server:
	go run main.go
mock:#去 db/sqlc 包里找到名为 Store 的接口，自动生成一个实现了这个接口的 mock 结构体（通常叫 MockStore）。
	mockgen -package mockdb -destination db/mock/store.go github.com/techschool/simplebank/db/sqlc Store
.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test server mock#Makefile的保护写法
