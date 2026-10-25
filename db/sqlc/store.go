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
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err:%v,rb err:%v", err, rbErr)
		}
	}
	return tx.Commit()
}

//T

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
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{ //新建一个交易Transfer
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
		//TODO: 更新账户的余额

		return nil
	}) //execTx的第二个参数是函数

	return result, err
}
