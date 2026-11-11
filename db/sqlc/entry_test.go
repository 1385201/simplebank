package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/techschool/simplebank/util"
)

/*
func createRandomEntry(t *testing.T) Entry {
	user := createRandomUser(t)
	arg1 := CreateAccountParams{
		Owner:    user.Username,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	account, err := testQueries.CreateAccount(context.Background(), arg1)
	require.NoError(t, err)

	arg2 := CreateEntryParams{
		AccountID: account.ID,
		Amount:    util.RandomMoney(),
	}
	entry, err := testQueries.CreateEntry(context.Background(), arg2)
	require.NoError(t, err)

	require.NotEmpty(t, entry)

	require.NotZero(t, entry.ID)
	require.Equal(t, account.ID, entry.AccountID)
	require.Equal(t, arg2.Amount, entry.Amount)
	require.NotZero(t, entry.CreatedAt)

	return entry

}
func TestCreateEntry(t *testing.T) {
	createRandomEntry(t)
}

func TestGetEntry(t *testing.T) {
	entry1 := createRandomEntry(t)
	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)

	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second)
}

func TestListEntries(t *testing.T) {
	/*
		user := createRandomUser(t)
		arg1 := CreateAccountParams{
			Owner:    user.Username,
			Balance:  util.RandomMoney(),
			Currency: util.RandomCurrency(),
		}
		account, err := testQueries.CreateAccount(context.Background(), arg1)
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			arg2 := CreateEntryParams{
				AccountID: account.ID,
				Amount:    util.RandomMoney(),
			}
			_, err = testQueries.CreateEntry(context.Background(), arg2)
			require.NoError(t, err)
		}
		arg3 := ListEntriesParams{
			AccountID: account.ID,
			Limit:     10,
			Offset:    0,
		}
		entries, err := testQueries.ListEntries(context.Background(), arg3)
		require.NoError(t, err)

		require.Len(t, entries, 10)
		for _, entry := range entries {
			require.NotEmpty(t, entry)
			require.NotZero(t, entry.ID)
			require.Equal(t, account.ID, entry.AccountID)
			//require.NotZero(t, entry.Amount)
			require.NotZero(t, entry.CreatedAt)
		}
*/
/*var lastentry Entry
	for i := 0; i < 10; i++ {
		lastentry = createRandomEntry(t)
	}

	arg := ListEntriesParams{
		AccountID: lastentry.AccountID,
		Limit:     5,
		Offset:    0,
	}
	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	for _, entry := range entries {
		require.NotEmpty(t, entry)
		require.NotZero(t, entry.ID)
		require.Equal(t, arg.AccountID, entry.AccountID)
		require.NotZero(t, entry.CreatedAt)
	}

}
*/

func createRandomEntry(t *testing.T, account Account) Entry {
	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount:    util.RandomMoney(),
	}

	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}
func TestCreateEntry(t *testing.T) {
	account := createRandomAccount(t)
	createRandomEntry(t, account)
}
func TestGetEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry1 := createRandomEntry(t, account)
	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second)
}
func TestListEntries(t *testing.T) {
	account := createRandomAccount(t)
	for i := 0; i < 10; i++ {
		createRandomEntry(t, account)
	}

	arg := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
		require.Equal(t, arg.AccountID, entry.AccountID)
	}
}
