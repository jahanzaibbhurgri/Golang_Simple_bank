//orm//
GORM: difficult as we have to learn the sql queries and it runs very slowly on high load
SQLX: convientent to use but still not approachable as failure wont occur until runtime
SQLC: APPROACHABLE: as we pass the queries and we can the built in generated code, automatic code generation
catch sql query before generating code, and full support of postgres but not of sql 

*but thou sqlx is widely used but in this sqlc because of postgres*

*Database trasaction*
1) to provide a reliable and consistent unit of work,even in case of system failure
2) to provide the isolation between programs that access the database concurrently do understand this by the example

both of these properties must satisfy the acid properties
1) atomicity: either all of the operations complete successfully or the transaction fails and db is unchanged
2) consistency: the db state must be valid after the transaction.All 
constraints must be satisfied
3) isolation: current trasaction must not affect each other
4) durability: data written by the successful must be recorded in a persistant storage

how to run the sql tx?
begin;   begin;
commit;   rollback;


Why Closing errs and results Is Unnecessary?
In Go, channels should be closed only by the sender when no more values will be sent. However, in your case:

Each goroutine (worker) sends exactly one error and one result.
The main goroutine (TestTransferTx) is the only receiver.
The main function is always reading exactly n values from both channels, ensuring all data is processed.
Since the receiver (TestTransferTx) knows exactly how many messages to expect (n), closing the channels is not required. The function will read n messages, then automatically exit.

Key Rule: A receiver can still read all sent values even if the channel is not closed. Closing is mainly useful when multiple receivers are waiting, so they know when to stop.

When Should You Close a Channel?

If multiple goroutines are waiting on a channel, and they need to be notified that no more values will come.
If you are using range to iterate over the channel, because range stops only when the channel is closed.

🔹 Example where closing is necessary:
results := make(chan int)

// Worker goroutine
go func() {
    for i := 1; i <= 5; i++ {
        results <- i
    }
    close(results) // Closing required, otherwise `range` below would block forever
}()

// Receiver using `range`
for res := range results {
    fmt.Println(res) // Prints numbers 1 to 5
}
✅ Closing ensures range stops, preventing a deadlock.


/* how the deadlock occurs in the multiple transaction and we solved it by using no key in the update query in the account and before that we for update in the front of the select statement for lock purpose */

//will do seperate usecases for that for the practice purpose and 
//make it clear 

You have two transactions (tx 1 and tx 2) that are trying to update accounts as part of a transfer process.

Here’s the sequence of operations:

Transaction 2 (tx 2) starts first:
tx 2 create transfer
tx 2 create entry 1
tx 2 create entry 2
tx 2 get account 1 for update (Locks account 1) ✅
tx 2 update account 1 (Still holding lock on account 1) ✅
tx 1 create transfer (tx 1 starts)
tx 2 get account 2 (Locks account 2) ✅
tx 2 update account 2 (Still holding lock on account 2) ✅
Transaction 1 (tx 1) starts:
tx 1 create entry 1
tx 1 create entry 2
tx 1 get account 1 for update (Tries to lock account 1, but tx 2 already locked it) ❌ (Blocked)
tx 1 update account 1 (Still waiting)
tx 1 get account 2 (Tries to lock account 2, but tx 2 already locked it) ❌ (Blocked)
At this point, both transactions are waiting on each other:

tx 1 is waiting for tx 2 to release account 1.
tx 2 is waiting for tx 1 to release account 2.
Since neither transaction can proceed, PostgreSQL detects a deadlock and cancels one of them.

How to Fix the Deadlock?
The main problem is that transactions lock the accounts in different orders:

tx 2 locks account 1 first, then account 2.
tx 1 locks account 2 first, then account 1.
The solution is to make sure that all transactions always lock accounts in the same order.

Option 1: Always Lock in Ascending Order
Change your logic so that whenever a transfer happens, you always lock the accounts in the same order (e.g., by sorting them by id):
SELECT * FROM accounts WHERE id IN (:account_1, :account_2) ORDER BY id FOR UPDATE;
This way:

If account_1 is smaller, it gets locked first.
If account_2 is smaller, it gets locked first.
No two transactions will try to lock accounts in a different order.

Option 2: Use SERIALIZABLE Isolation Level
Another way to avoid deadlocks is by using serializable transactions, which ensures that transactions run one after another instead of in parallel:


SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
However, this may reduce performance since transactions won’t run at the same time.

Option 3: Retry the Transaction on Deadlock
If a deadlock happens, retrying the transaction can help. Many databases provide retry mechanisms for handling deadlocks.

For example, in PostgreSQL:

typescript
Copy
Edit
async function executeWithRetry(transactionFn, retries = 3) {
    for (let i = 0; i < retries; i++) {
        try {
            return await transactionFn();
        } catch (error) {
            if (error.code === '40001') { // Deadlock error code
                console.log('Deadlock detected, retrying...');
            } else {
                throw error;
            }
        }
    }
}
This will retry the transaction if a deadlock occurs.

Using FOR NO KEY UPDATE instead of FOR UPDATE helped prevent deadlocks because it reduces the lock severity. I’ll explain why that worked and also explain your foreign key constraints.

1. Why Did FOR NO KEY UPDATE Prevent Deadlocks?
In PostgreSQL, different types of locks exist when updating data:

Lock Type	What It Does	Can It Block Other Transactions?
FOR UPDATE	Strongest lock: prevents other transactions from updating or locking the row	Yes, can cause deadlocks if locks are acquired in different orders
FOR NO KEY UPDATE	Weaker lock: prevents updates but allows foreign key constraints and some updates	Less likely to cause deadlocks
How It Fixed Your Deadlock Issue
Before (FOR UPDATE):

tx 1 locks account 1 fully.
tx 2 locks account 2 fully.
Then, they try to access each other’s locked accounts → deadlock.
After (FOR NO KEY UPDATE):

tx 1 gets a weaker lock on account 1.
tx 2 gets a weaker lock on account 2.
PostgreSQL allows them to proceed without fully blocking each other.
Thus, FOR NO KEY UPDATE helps reduce contention without completely blocking access.

