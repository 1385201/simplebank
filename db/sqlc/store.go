package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	*Queries         //用来执行普通数据库操作
	db       *sql.DB //用来开事务
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error { //事务执行模板,(接收一个函数作为参数，函数 fn，就是你要执行的具体业务逻辑)

	tx, err := store.db.BeginTx(ctx, nil) //告诉数据库，我要开始一段事务，接下来的所有操作，都在这个通道里进行，nil表示默认隔离级别，ctx表示流水线工人的传话筒
	if err != nil {
		return err
	}
	q := New(tx)    //q是里面装着*sql.Tx的*Queries类型
	err = fn(q)     //去执行传入的函数
	if err != nil { //如果err不是nil，说明刚刚执行的fn函数失败了，此时需要回滚到之前的状态
		if rbErr := tx.Rollback(); rbErr != nil { //tx.Rollback()执行回滚操作，并将错误值返回给rbErr,如果rbERR！=nil说明回滚操作也失败了，此时只能错误打印出来
			return fmt.Errorf("tx err:%v,rb err:%v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

// 转账所需的参数
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"` // 从哪个账户转出       //`json:"from_account_id"`当把这个结构体转换成 JSON 数据时，请使用我指定的这个名字。
	ToAccountID   int64 `json:"to_account_id"`   // 转到哪个账户去
	Amount        int64 `json:"amount"`          // 转多少钱
}

// 转账的结果
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`     // 转账流水记录
	FromAccount Account  `json:"from_account"` // 转出方的最新账户信息
	ToAccount   Account  `json:"to_account"`   // 转入方的最新账户信息
	FromEntry   Entry    `json:"from_entry"`   // 转出方的账单明细
	ToEntry     Entry    `json:"to_entry"`     // 转入方的账单明细
}

// 转账函数
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) { //两个结构体分别是转账函数的输入和输出
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{ //新建一个交易Transfer，但是因为此处的q结构体里装的db是*sql.Tx类型,是执行一条像CreateTransfer这样的SQL语句，数据库不立刻保存的
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{ //新建转出方的账单
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{ //新建转入方的账单
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		//更新account1
		account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}
		result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.FromAccountID,
			Balance: account1.Balance - arg.Amount,
		})
		if err != nil {
			return err
		}
		//更新account2
		account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
		if err != nil {
			return err
		}
		result.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.ToAccountID,
			Balance: account2.Balance + arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	}) //execTx的第二个参数是函数

	return result, err
}
