package db

import (
	"context"
	_ "database/sql"
	"simplebank/util"
	"testing"
	_ "time"

	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T, fromAccount Account, toAccount Account) Transfer {
	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        util.RandomMoney(),
	}
	transfer, err := testStore.CreateTransfer(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.NotZero(t, transfer.ID)
	require.Equal(t, fromAccount.ID, transfer.FromAccountID)
	require.Equal(t, toAccount.ID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	createRandomTransfer(t, account1, account2)
}

func TestGetTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	transfer1 := createRandomTransfer(t, account1, account2)

	transfer2, err := testStore.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.Equal(t, transfer1.CreatedAt, transfer2.CreatedAt)
}

func TestListTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	for i := 0; i < 10; i++ {
		createRandomTransfer(t, account1, account2)
	}
	arg := ListTransfersFromToParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Limit:         5,
		Offset:        5,
	}
	transfers, err := testStore.ListTransfersFromTo(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, transfers)

	require.Len(t, transfers, 5)
	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}
