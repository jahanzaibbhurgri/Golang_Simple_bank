package db

import (
	"context"
	"database/sql"
	"fmt"
)

var txKey = struct{}{}

// This is not a key-value pair data structure, 
// but rather a unique, zero-sized value. 
// It is often used as a unique key in contexts 
// where you need a unique identifier,
//  such as in the context.WithValue function.

//store provides all functions to execute db queries and transactions
type Store struct {
	*Queries  //this is the way called composition(extended the functioanlity of Queries) 
    db *sql.DB 
}

//NewStore creates a new store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
		Queries: New(db),
	}
}

//execTx executes a function within a database transaction
func (store *Store) execTx(ctx context.Context,fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx,nil)
	if err != nil {
		return err
	}
	
	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}
	
	return tx.Commit()
}

type TransferTxParams struct {
  FromAccountID int64 `json:"from_account_id"`
  ToAccountID int64 `json:"to_account_id"`
  Amount int64 `json:"amount"`
}

type TransferTxResult struct {
  Transfer Transfer `json:"transfer"`
  FromAccount Account `json:"from_account"`
  ToAccount  Account `json:"to_account"`
  FromEntry  Entry `json:"from_entry"`
  ToEntry Entry `json:"to_entry"`
  
}


// TransferTx performs a money transfer from one account to another
// It creates a transfer record, add account entries, and update accounts' balance within a single database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
    var result TransferTxResult

    err := store.execTx(ctx, func(q *Queries) error {
        var err error

        txName := ctx.Value(txKey)
        
        fmt.Println(txName, "create transfer")
        result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
            FromAccountID: arg.FromAccountID,
            ToAccountID:   arg.ToAccountID,
            Amount:        arg.Amount,
        })
        if err != nil {
            return err
        }
        fmt.Println(txName, "create entry 1")
        result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
            AccountID: int32(arg.FromAccountID),
            Balance:    -arg.Amount,
        })
        if err != nil {
            return err
        }
        
        fmt.Println(txName, "create entry 2")
        result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
            AccountID: int32(arg.ToAccountID),
            Balance:    arg.Amount,
        })
        if err != nil {
            return err
        }
        
		fmt.Println(txName, "get account 1 and update it ")
		result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:      arg.FromAccountID,
			Amount:  -arg.Amount,
		})
		if err != nil {
			return err
		   }

        fmt.Println(txName, "get account 2 and update it")
		result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:      arg.ToAccountID,
			Amount:  arg.Amount,
		})
		if err != nil {
			return err
		   }
		
   
        return nil
    })

	//getting
    return result, err
}

//what is the closure function in go//
//clousre is often used when we used to get the results from the callback function//
//go lacks supports for the generic type//