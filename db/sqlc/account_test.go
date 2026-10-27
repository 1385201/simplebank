package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/techschool/simplebank/util"
)

func createRandomAccount(t *testing.T) Account {
	// 第1步：准备一个"随机账户"的数据
	arg := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	// 第2步：把这个数据真正写进数据库
	account, err := testQueries.CreateAccount(context.Background(), arg)
	//context.Background()：是给数据库方法的“通行证”，表示“无超时、不取消，正常干”。
	//require：是测试的“保安”。检查合格就放行，检查出致命错误就中断测试

	// 第3步：检查
	require.NoError(t, err) // 检查：有没有报错？有就失败
	/*
			// require.NoError 的伪代码
		func NoError(t *testing.T, err error) {
		    if err != nil { // 如果 err 不是空的（说明真的报错了）
		        t.FailNow() // 拿起遥控器，按下“拉闸停电”按钮！
		    }
		}
	*/
	require.NotEmpty(t, account) // 检查：返回的东西是不是空的？

	require.Equal(t, arg.Owner, account.Owner)       // 检查：名字对不对
	require.Equal(t, arg.Balance, account.Balance)   // 检查：金额对不对
	require.Equal(t, arg.Currency, account.Currency) // 检查：币种对不对

	require.NotZero(t, account.ID)        // 检查：数据库有没有生成 ID（ID 不能是 0）
	require.NotZero(t, account.CreatedAt) // 检查：有没有生成创建时间
	return account
}

func TestCreateAccount(t *testing.T) {
	// ① 调用辅助函数 createRandomAccount(t)
	//    这个函数内部已经完成了"创建账户 + 逐字段断言"的完整流程：
	//      - 随机生成 Owner / Balance / Currency
	//      - 调用 testQueries.CreateAccount 写进数据库
	//      - 用 require.XXX 断言：没报错、返回不为空、
	//        写进去的字段和返回的字段一致、ID 和 CreatedAt 已生成
	//
	//    返回值 account 被丢弃了,因为 TestCreateAccount 只需要"验证创建功能对不对",
	//    不需要拿这个账户去做后续操作（不像 TestGetAccount 需要 account1.ID）
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	// ① 先随机创建一个账户,返回它的完整信息
	//    account1 = "当初放进数据库的那个账户"（我们手里有它的所有字段）
	account1 := createRandomAccount(t)

	// ② 用 account1 的 ID 去数据库里查询
	//    account2 = "从数据库里查出来的那个账户"
	//    context.Background() = 不设超时/取消,正常查
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)

	// ③ 断言：查询操作本身不能报错
	require.NoError(t, err)

	// ④ 断言：查出来的 account2 不能是空对象
	require.NotEmpty(t, account2)

	// ⑤ 断言：查出来的 ID 必须和放进去的 ID 一致
	require.Equal(t, account1.ID, account2.ID)

	// ⑥ 断言：Owner（名字）一致
	require.Equal(t, account1.Owner, account2.Owner)

	// ⑦ 断言：Balance（余额）一致
	require.Equal(t, account1.Balance, account2.Balance)

	// ⑧ 断言：Currency（币种）一致
	require.Equal(t, account1.Currency, account2.Currency)

	// ⑨ 断言：CreatedAt（创建时间）一致
	//    用 WithinDuration 而不是 Equal,允许 1 秒误差
	//    原因：数据库时间精度可能和 Go 的 time.Time 不完全一致,
	//          直接 Equal 可能因微小差异"假失败"
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second) //(理论上是一样的，只是数据库的精度和go的精度有点不同，所以用time.second来运行一定的误差)
}
func TestUpdateAccount(t *testing.T) {
	// ① 先随机创建一个账户,返回它的完整信息（含自动生成的 ID）
	//    这个账户就是我们要"改"的对象,account1 代表"改之前的状态"
	account1 := createRandomAccount(t)

	// ② 准备"更新参数"：
	//    - ID：要改哪个账户（用刚建好的 account1.ID）
	//    - Balance：改成一个新的随机金额（这就是我们想改的字段）
	arg := UpdateAccountParams{
		ID:      account1.ID,
		Balance: util.RandomMoney(),
	}
	// ③ 真正执行更新,account2 是数据库返回的"改之后"的账户
	account2, err := testQueries.UpdateAccount(context.Background(), arg)
	// ④ 断言：更新操作本身不能报错
	require.NoError(t, err)
	// ⑤ 断言：返回的 account2 不能是空对象
	require.NotEmpty(t, account2)
	// ⑥ 断言：ID 没变（更新不能改主键）
	require.Equal(t, account1.ID, account2.ID)
	// ⑦ 断言：Owner 没变（我们没打算改名字,它必须原样保留）
	require.Equal(t, account1.Owner, account2.Owner)
	// ⑧ 断言：Balance 变成了我们设定的新值（这是"更新成功"的核心证据）
	require.Equal(t, arg.Balance, account2.Balance)
	// ⑨ 断言：Currency 没变（没打算改币种,必须原样保留）
	require.Equal(t, account1.Currency, account2.Currency)
	// ⑩ 断言：CreatedAt 没变,允许 1 秒误差（时间精度问题）
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
}

func TestDeleteAccount(t *testing.T) {
	// ① 先随机创建一个账户,返回它的完整信息（含自动生成的 ID）
	//    这个账户就是我们要"删"的对象,account1 代表"删之前还存在的状态"
	account1 := createRandomAccount(t)

	// ② 真正执行删除：把 account1 的 ID 传给删除方法
	//    注意这里必须接住返回值（account1.ID）,因为要告诉数据库删哪个
	err := testQueries.DeleteAccount(context.Background(), account1.ID)

	// ③ 断言：删除操作本身不能报错
	//    如果 err != nil,说明 DELETE 语句执行失败（SQL 错、连接断等）
	require.NoError(t, err)

	// ④ 再用同一个 ID 去查一次,account2 是"删之后"查询的结果（期望查不到）
	//    这是验证删除是否真的生效的关键一步
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)

	// ⑤ 断言：这次查询「必须报错」
	//    注意这里是 require.Error,不是 NoError！
	//    逻辑：账户已删,再查就应该"查不到"；如果没报错,说明删了还在 = bug
	require.Error(t, err)

	// ⑥ 断言：报的错「必须正好是 sql.ErrNoRows」（查无此行）
	//    用 EqualError 精确比对错误字符串,确保不是别的错误
	//    （比如万一报的是"连接超时",那测试就无效了）
	require.EqualError(t, err, sql.ErrNoRows.Error())

	// ⑦ 断言：返回的 account2 必须是空对象
	//    既然查不到,就不该返回任何数据（双重保险）
	require.Empty(t, account2)
}

func TestListAccounts(t *testing.T) {
	// ① 循环创建 10 个随机账户,为分页查询准备足够的数据
	//    这里故意用 for 循环而不是手动写 10 次
	//    返回值被忽略了,因为分页测试只关心"数据库里有多少条",
	//    不需要拿到每个账户的细节去比对
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	// ② 准备"分页参数"：
	//    - Limit=5 ：每页取 5 条
	//    - Offset=5：跳过前 5 条（相当于"取第 2 页"）
	//    组合起来就是："跳过前 5 条,从第 6 条开始取 5 条"
	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	// ③ 真正执行分页查询,accounts 是查回来的账户列表（一个切片）
	accounts, err := testQueries.ListAccounts(context.Background(), arg)

	// ④ 断言：查询操作本身不能报错
	require.NoError(t, err)

	// ⑤ 断言：返回的列表长度必须正好是 5
	//    这是分页测试的核心：要 5 条就必须给 5 条
	//    多了或少了都说明 LIMIT/OFFSET 逻辑有 bug
	require.Len(t, accounts, 5)

	// ⑥ 遍历返回的每一条账户,逐条检查
	for _, account := range accounts {
		// ⑦ 断言：每个账户对象都不能是空的
		//    防止查询返回了"空结构体"（虽然理论上不会,
		//    但加上这层检查更严谨）
		require.NotEmpty(t, account)
	}
}
