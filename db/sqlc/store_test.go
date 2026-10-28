package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)          //因为TransferTx是绑在*store上的方法，所以需要先搞个*Store类型的值出来，才能测试
	account1 := createRandomAccount(t) //同一个包里可以调用未导出函数
	account2 := createRandomAccount(t)

	fmt.Println(">> before:", account1.Balance, account2.Balance)

	//假设同时处理5个事务
	n := 5
	amount := int64(10)

	//两个通 道，连接同时执行的go程序，并允许他们在没有明确锁定的情况下安全的相互共享数据（不同的流水线工人的对讲机）
	errs := make(chan error)               //接收错误的通道，chan error是类型
	results := make(chan TransferTxResult) //接收TransferTxResult数据的通道，chan TransferTxResult也是类型

	for i := 0; i < n; i++ { //模拟高并发场景，测试你的“转账事务”在同时发生 5 次时，会不会出 Bug

		go func() { //go 关键字表示开启一个 Goroutine（协程）,for 循环会瞬间执行 5 次,每一次都会丢出一个协程去后台执行。这 5 次转账是“同时”发生的，谁也不等谁。
			ctx := context.Background()
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err //通道是五个子线程共享的，具有严格的并发安全，如果多个子线程往里面放数据，会排队，而不是冲突
			results <- result

		}()
	}
	//检查结果(检查每一次动作是否合法)
	existed := make(map[int]bool)
	for i := 0; i < n; i++ { //轮流检查五次子协程的结果正确性
		err := <-errs //从errs通道里拿结果，如果没有就等待
		require.NoError(t, err)

		result := <-results //从results通道里拿结果，如果没有就等待

		require.NotEmpty(t, result)

		//检查transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID) //确保transfer记录确实在数据库中
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

		//检查账户
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)              // 账户信息不能为空
		require.Equal(t, account1.ID, fromAccount.ID) // 账户 ID 必须和预期的转出方 ID 一致

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)              // 账户信息不能为空
		require.Equal(t, account2.ID, toAccount.ID) // 账户 ID 必须和预期的转入方 ID 一致

		// 检查账户余额
		fmt.Println(">> tx:", fromAccount.Balance, toAccount.Balance)
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2) // 转出方减少的钱，必须等于转入方增加的钱
		require.True(t, diff1 > 0)     // 转出的金额必须是正数（大于0）
		// 减少的金额必须是 amount 的整数倍 (即 1 * amount, 2 * amount, 3 * amount, ..., n * amount)
		require.True(t, diff1%amount == 0)
		k := int(diff1 / amount)
		// 验证 k 的范围是否合法：至少是第 1 次转账，最多是第 n 次转账
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k) //检查existed[k]是否为true,如果为true，说明重复执行了，代表数据库并发出大 bug 了
		existed[k] = true
	}

	// 检查最终的结果是否正确（上面只保证了转账的过程没有出错）

	// 1. 去数据库查询转出方账户的最新真实状态
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err) // 确保查询过程没有报错

	// 2. 去数据库查询转入方账户的最新真实状态
	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err) // 确保查询过程没有报错

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)

	// 3. 验证转出方总账：
	// 初始余额 - (并发次数 n * 单次金额 amount) = 数据库里的最终真实余额
	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)

	// 4. 验证转入方总账：
	// 初始余额 + (并发次数 n * 单次金额 amount) = 数据库里的最终真实余额
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)

}
