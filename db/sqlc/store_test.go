package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)          //因为TransferTx是绑在*store上的方法，所以需要先搞个*Store类型的值出来，才能测试
	account1 := createRandomAccount(t) //同一个包里可以调用未导出函数
	account2 := createRandomAccount(t)

	//假设同时处理5个事务
	n := 5
	amount := int64(10)

	//通道，连接同时执行的go程序，并允许他们在没有明确锁定的情况下安全的相互共享数据（不同的流水线工人的对讲机）
	errs := make(chan error)               //接收错误的通道
	results := make(chan TransferTxResult) //接收TransferTxResult数据的通道

	for i := 0; i < n; i++ { //模拟高并发场景，测试你的“转账事务”在同时发生 5 次时，会不会出 Bug
		go func() { //go 关键字表示开启一个 Goroutine（协程）,for 循环会瞬间执行 5 次,每一次都会丢出一个协程去后台执行。这 5 次转账是“同时”发生的，谁也不等谁。
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result

		}()
	}
	//检查结果
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		//检查transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetAccount(context.Background(), transfer.ID) //确保transfer记录确实在数据库中
		require.NoError(t, err)

		//检查entry1:fromEntry
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, amount, -fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		//检查entry2:toEntry
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		//TODO:检查两个账户的变动

	}

}
