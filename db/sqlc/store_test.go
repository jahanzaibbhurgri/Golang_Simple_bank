package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)
	fmt.Println(">> Before:", account1.Balance, account2.Balance)

	// Run n concurrent transfer transactions
	n := 5
	amount := int64(10)

	errs := make(chan error, n) // Fixed: Correct channel name and buffer size
	results := make(chan TransferTxResult, n)

	for i := 0; i < n; i++ { // Fixed: Changed `<=` to `<` to avoid extra loop iteration
        txName := fmt.Sprintf("tx %d", i+1) //we are using this to debug which query is being called for the deadlock situation//
		// It formats a string according to a format 
		// specifier and returns the formatted string.

		go func() {
            ctx := context.WithValue(context.Background(), txKey, txName)
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			results <- result // Send result to channel
			errs <- err       // Send error to channel
		}()
	}

	//check existed//
	existed := make(map[int]bool)

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
		_, err = store.GetTransfer(context.Background(), transfer.ID) //do understand this//
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

		//check accounts//
		//this is where the updateAccounts test cases is created//
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		//check account balance//
		fmt.Println(">> After:", fromAccount.Balance, toAccount.Balance)
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) // 1 * amount2 * amount3 * ... n * amount

		//   Why Do We Need This Check?
		//   When performing n concurrent transactions, each transaction transfers a fixed amount.
		//   Thus, after n transactions, the total transferred amount must be:

		//   Total Transferred
		//   =
		//   𝑘
		//   ×
		//   amount
		//   Total Transferred=k×amount
		//   where
		//   𝑘
		//   k is the number of successful transactions (an integer).

		//   If this check fails, it means:

		//   Some transactions were not executed properly.
		//   An inconsistent amount was deducted from account1.
		//   📌 Example 1: Successful Case
		//   🟢 Scenario: 5 Successful Transfers (each of 10)
		//   Transaction	Deducted from A	Added to B
		//   1st	-10	+10
		//   2nd	-10	+10
		//   3rd	-10	+10
		//   4th	-10	+10
		//   5th	-10	+10
		//   Total	-50	+50
		//   diff1
		//   =
		//   50
		//   diff1=50
		//   ✅ Check:

		//   require.True(t, diff1 % amount == 0) // 50 % 10 == 0 ✅ PASS

		/////////////////////////////
		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true

		//check the final update balances
		updateAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
		require.NoError(t, err)

		updateAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
		require.NoError(t, err)

		fmt.Println(">> After:", updateAccount1.Balance, updateAccount2.Balance)
		require.Equal(t, account1.Balance-amount*int64(n), updateAccount1.Balance)
		require.Equal(t, account2.Balance+amount*int64(n), updateAccount2.Balance)

	}

}

//here we using channels to get the data from the go routines//
// Alternative: Instead of using channels, you could also use a sync.WaitGroup to explicitly wait for all goroutines to finish before proceeding.

// Example using sync.WaitGroup:
// var wg sync.WaitGroup
// wg.Add(n) // Set the counter to `n`

// for i := 0; i < n; i++ {
//     go func() {
//         defer wg.Done() // Decrease counter when goroutine finishes
//         result, err := store.TransferTx(context.Background(), TransferTxParams{
//             FromAccountID: account1.ID,
//             ToAccountID:   account2.ID,
//             Amount:        amount,
//         })
//         results <- result
//         errs <- err
//     }()
// }

// wg.Wait() // Blocks until all goroutines finish

// // Read results after all goroutines complete
// for i := 0; i < n; i++ {
//     err := <-errs
//     require.NoError(t, err)

//     result := <-results
//     require.NotEmpty(t, result)
// }

// to debug the deadlock situation//
//we would be tdd approach first write the test cases which break the code//
//slowly improve the code to pass the test cases//
