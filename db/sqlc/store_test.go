package db

import (
	"context"
	"testing"
	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
    store := NewStore(testDB)

    account1 := CreateRandomAccount(t)
    account2 := CreateRandomAccount(t)

    // Run n concurrent transfer transactions
    n := 5
    amount := int64(10)

    errs := make(chan error,n) // Fixed: Correct channel name and buffer size
    results := make(chan TransferTxResult,n)

    for i := 0; i < n; i++ { // Fixed: Changed `<=` to `<` to avoid extra loop iteration
        go func() {
            result, err := store.TransferTx(context.Background(), TransferTxParams{
                FromAccountID: account1.ID,
                ToAccountID:   account2.ID,
                Amount:        amount,
            })

            results <- result // Send result to channel
            errs <- err       // Send error to channel
        }()
    }

    for i := 0; i < n; i++ { // Fixed: Changed `<=` to `<` and correctly read from channels
        err := <-errs
        require.NoError(t, err)

        result := <-results
        require.NotEmpty(t, result)
       
		//check transfer//
		transfer := result.Transfer
        require.NotEmpty(t, transfer)
        require.NotEmpty(t, transfer.FromAccountID)
        require.NotEmpty(t, transfer.ToAccountID)
        require.NotEmpty(t, transfer.Amount)
        require.NotEmpty(t, transfer.CreatedAt)

		   // Fetch and check transfer
		   _ , err = store.GetTransfer(context.Background(), transfer.ID) //do understand this//
		   require.NoError(t, err)


		   // Check FromEntry
		   fromEntry := result.FromEntry
		   require.NotEmpty(t, fromEntry)
		   require.Equal(t, account1.ID, int64(fromEntry.AccountID))
		   require.Equal(t, -amount, int64(fromEntry.Balance))
		   require.NotZero(t, fromEntry.ID)
		   require.NotZero(t, fromEntry.CreatedAt)
	   
		   _, err = store.GetEntry(context.Background(), fromEntry.ID)
		   require.NoError(t, err)
	   
		   // Check ToEntry
		   toEntry := result.ToEntry
		   require.NotEmpty(t, toEntry)
		   require.Equal(t, account2.ID, int64(toEntry.AccountID))
		   require.Equal(t, amount, int64(toEntry.Balance))
		   require.NotZero(t, toEntry.ID)
		   require.NotZero(t, toEntry.CreatedAt)
	   
		 _, err = store.GetEntry(context.Background(), toEntry.ID)
		  require.NoError(t, err)
    }
   

 
}
