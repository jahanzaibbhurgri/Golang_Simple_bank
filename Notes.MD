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

//isolation ways//
1) dirty read: a transaction reads data written by other uncommited transaction

example: there is 2 acc having 100 each 
update acc 1 with subtraction of 10 
and in second terminal select * form accounts where id = 1
so its value would be 90 thou it is not commited so this dirty read


2)non repeatable read: a transaction reads the same row twice and sees different value because of it has been modified by other commited transaction

3) phantom read: a transaction reexecutes a query to find rows that 
satisfy a condition and sees a different set of rows, due to changes by 
the other commited transaction

4) serialization anomaly: the result of group of concurrent 
commited transaction is impossible to achieve if we try to run them sequentially in any order without overlapping

Transaction Isolation Levels & Their Effects
Let's assume we have a bank account table:

id	balance
1	100
2	100

Now, two transactions are running simultaneously in different sessions. We'll go through each isolation level with examples and see how different anomalies are prevented.

1) Read Uncommitted (Lowest isolation, allows dirty reads)
👉 Can see uncommitted changes from other transactions.

Example (Dirty Read)
Session 1:

BEGIN TRANSACTION;
UPDATE accounts SET balance = 90 WHERE id = 1;
-- (Transaction not committed yet)

Session 2 (before Session 1 commits):
t
SELECT balance FROM accounts WHERE id = 1;
Result: 90 (Even though the update is not committed yet—this is a dirty read.)
If Session 1 does a rollback:

ROLLBACK;

Now the actual balance remains 100, but Session 2 has already seen 90, which was never committed.
Prevention
❌ Dirty Read NOT prevented
❌ Non-repeatable Read NOT prevented
❌ Phantom Read NOT prevented
❌ Serialization Anomaly NOT prevented

2) Read Committed (Prevents dirty reads)
👉 Only sees committed changes from other transactions.

Example (Prevents Dirty Read but allows Non-Repeatable Read)
Session 1:


BEGIN TRANSACTION;
UPDATE accounts SET balance = 90 WHERE id = 1;
-- (Transaction not committed yet)
Session 2 (before Session 1 commits):


SELECT balance FROM accounts WHERE id = 1;
Result: 100 (because it only reads committed data)

After Session 1 commits:

COMMIT;
Session 2 (Re-executing same query):

SELECT balance FROM accounts WHERE id = 1;
Result: 90 (The value has changed, leading to a non-repeatable read.)

Prevention
✅ Dirty Read Prevented
❌ Non-repeatable Read NOT prevented
❌ Phantom Read NOT prevented
❌ Serialization Anomaly NOT prevented

3) Repeatable Read (Prevents dirty reads & non-repeatable reads)
👉 Same read query always returns the same result within a transaction.

Example (Prevents Non-Repeatable Read but allows Phantom Read)
Session 1:

BEGIN TRANSACTION;

SELECT balance FROM accounts WHERE id = 1;
Result: 100

Session 2 (Modifies the same row and commits):

BEGIN TRANSACTION;
UPDATE accounts SET balance = 90 WHERE id = 1;
COMMIT;
Session 1 (Re-executing same query):

SELECT balance FROM accounts WHERE id = 1;
Result: 100 (Not affected by committed changes—prevents non-repeatable reads.)

Session 2 (Inserts a new row and commits):

INSERT INTO accounts VALUES (3, 50);
COMMIT;
Session 1 (Now runs a query for all accounts with balance > 0):

SELECT * FROM accounts WHERE balance > 0;
Result: Includes ID 3, which wasn’t there before (Phantom Read).

Prevention
✅ Dirty Read Prevented
✅ Non-repeatable Read Prevented
❌ Phantom Read NOT prevented
❌ Serialization Anomaly NOT prevented

4) Serializable (Highest isolation, prevents everything)
👉 Transactions execute as if they were run sequentially (one after another).

Example (Prevents Phantom Read)
Session 1 (Starts a transaction and reads all accounts):

BEGIN TRANSACTION;
SELECT * FROM accounts WHERE balance > 0;
Result: (IDs 1 & 2 only)

Session 2 (Tries to insert a new account and commit):
BEGIN TRANSACTION;
INSERT INTO accounts VALUES (3, 50);
COMMIT;
🚨 Blocked! Because Session 1 is running a serializable transaction, it prevents new rows from being inserted until it completes.

Session 1 (After committing its transaction, now Session 2 can proceed):
COMMIT;

Prevention
✅ Dirty Read Prevented
✅ Non-repeatable Read Prevented
✅ Phantom Read Prevented
✅ Serialization Anomaly Prevented

Summary Table
Isolation Level	Dirty Read	Non-Repeatable Read	Phantom Read	Serialization Anomaly
Read Uncommitted	❌ Not Prevented	❌ Not Prevented	❌ Not Prevented	❌ Not Prevented
Read Committed	✅ Prevented	❌ Not Prevented	❌ Not Prevented	❌ Not Prevented
Repeatable Read	✅ Prevented	✅ Prevented	❌ Not Prevented	❌ Not Prevented
Serializable	✅ Prevented	✅ Prevented	✅ Prevented	✅ Prevented

Differences in Isolation Behavior Between SQL (MySQL) & PostgreSQL

1) Dirty Read (Uncommitted Data Read)
✅ MySQL (Only in InnoDB with Read Uncommitted)

MySQL allows dirty reads when using READ UNCOMMITTED.
In READ COMMITTED, dirty reads are prevented.
✅ PostgreSQL (Always Prevented)

PostgreSQL never allows dirty reads, even at READ UNCOMMITTED.
PostgreSQL treats READ UNCOMMITTED as READ COMMITTED.

So, dirty reads are always prevented in PostgreSQL.
🔹 Example in MySQL (Dirty Read Possible)

-- Session 1 (Uncommitted Update)
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
BEGIN;
UPDATE accounts SET balance = 90 WHERE id = 1;
-- (Transaction NOT committed)
sql
Copy
Edit
-- Session 2 (Reads Uncommitted Data)
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
SELECT balance FROM accounts WHERE id = 1;
-- Output: 90 (Dirty Read, Even Though Not Committed)
🔹 Example in PostgreSQL (Dirty Read NOT Possible)

sql
Copy
Edit
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
BEGIN;
UPDATE accounts SET balance = 90 WHERE id = 1;
sql
Copy
Edit
-- Another Session
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
SELECT balance FROM accounts WHERE id = 1;
-- Output: 100 (Dirty Read prevented!)
🚀 PostgreSQL always prevents dirty reads, while MySQL allows them at READ UNCOMMITTED.


Differences in Isolation Behavior Between SQL (MySQL) & PostgreSQL
1) Dirty Read (Uncommitted Data Read)
✅ MySQL (Only in InnoDB with Read Uncommitted)

MySQL allows dirty reads when using READ UNCOMMITTED.
In READ COMMITTED, dirty reads are prevented.
✅ PostgreSQL (Always Prevented)

PostgreSQL never allows dirty reads, even at READ UNCOMMITTED.
PostgreSQL treats READ UNCOMMITTED as READ COMMITTED.
So, dirty reads are always prevented in PostgreSQL.
🔹 Example in MySQL (Dirty Read Possible)

sql
Copy
Edit
-- Session 1 (Uncommitted Update)
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
BEGIN;
UPDATE accounts SET balance = 90 WHERE id = 1;
-- (Transaction NOT committed)
sql
Copy
Edit
-- Session 2 (Reads Uncommitted Data)
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
SELECT balance FROM accounts WHERE id = 1;
-- Output: 90 (Dirty Read, Even Though Not Committed)
🔹 Example in PostgreSQL (Dirty Read NOT Possible)

sql
Copy
Edit
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
BEGIN;
UPDATE accounts SET balance = 90 WHERE id = 1;
sql
Copy
Edit
-- Another Session
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
SELECT balance FROM accounts WHERE id = 1;
-- Output: 100 (Dirty Read prevented!)
🚀 PostgreSQL always prevents dirty reads, while MySQL allows them at READ UNCOMMITTED.

2) Non-Repeatable Read (Different Data for the Same Query in the Same Transaction)
✅ MySQL (Issue in Read Committed, Prevented in Repeatable Read)

MySQL allows non-repeatable reads at READ COMMITTED.
Using REPEATABLE READ prevents this by keeping snapshot-based reads.
✅ PostgreSQL (Issue in Read Committed, Prevented in Repeatable Read)

PostgreSQL also allows non-repeatable reads at READ COMMITTED.
Using REPEATABLE READ prevents this in PostgreSQL as well.
🔹 Example in MySQL/PostgreSQL (READ COMMITTED Allows Non-Repeatable Read)

sql
Copy
Edit
-- Session 1 (Reads Before Update)
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;
BEGIN;
SELECT balance FROM accounts WHERE id = 1; -- Returns 100
sql
Copy
Edit
-- Session 2 (Commits Update)
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;
BEGIN;
UPDATE accounts SET balance = 90 WHERE id = 1;
COMMIT;
sql
Copy
Edit
-- Session 1 (Reads Again)
SELECT balance FROM accounts WHERE id = 1;
-- Output: 90 (Non-Repeatable Read, Value Changed!)
🚀 Both MySQL & PostgreSQL prevent this at REPEATABLE READ.

3) Phantom Read (New Rows Appear in the Same Transaction)
✅ MySQL (Allowed in Repeatable Read, Prevented in Serializable)

MySQL allows phantom reads at REPEATABLE READ because it only locks rows but not the whole table.
SERIALIZABLE prevents phantom reads by locking tables.
✅ PostgreSQL (Allowed in Repeatable Read, Prevented in Serializable)

PostgreSQL also allows phantom reads at REPEATABLE READ.
SERIALIZABLE in PostgreSQL prevents phantom reads using Serializable Snapshot Isolation (SSI) (different from MySQL).
🔹 Example in MySQL/PostgreSQL (REPEATABLE READ Allows Phantom Read)

sql
Copy
Edit
-- Session 1 (Reads List of Accounts)
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
BEGIN;
SELECT * FROM accounts WHERE balance > 50;
sql
Copy
Edit
-- Session 2 (Inserts New Row & Commits)
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
BEGIN;
INSERT INTO accounts VALUES (3, 70);
COMMIT;
sql
Copy
Edit
-- Session 1 (Reads Again)
SELECT * FROM accounts WHERE balance > 50;
-- Output: New row appears (Phantom Read occurred!)
🚀 In MySQL, SERIALIZABLE locks tables. In PostgreSQL, SERIALIZABLE prevents phantoms using SSI.

4) Serialization Anomaly (Concurrent Transactions Give Incorrect Results)
✅ MySQL (SERIALIZABLE Uses Table Locks, Performance Hit!)

MySQL enforces SERIALIZABLE by locking entire tables, blocking concurrent writes.
✅ PostgreSQL (SERIALIZABLE Uses Serializable Snapshot Isolation, No Locks!)

PostgreSQL uses Serializable Snapshot Isolation (SSI) instead of locking.
More scalable than MySQL because it doesn't block writes.
🔹 Example in MySQL (SERIALIZABLE Blocks Writes)

sql
Copy
Edit
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
BEGIN;
UPDATE accounts SET balance = balance - 10 WHERE id = 1;
-- (Any other transaction modifying the table is BLOCKED!)
🚀 PostgreSQL's SERIALIZABLE is better because it avoids full table locks!

Final Summary of MySQL vs PostgreSQL Isolation Handling
Isolation Level	MySQL Behavior	PostgreSQL Behavior
Read Uncommitted	Allows dirty reads	Same as Read Committed (No Dirty Reads)
Read Committed	Prevents dirty reads but allows non-repeatable reads	Prevents dirty reads but allows non-repeatable reads
Repeatable Read	Prevents non-repeatable reads but allows phantom reads	Prevents non-repeatable reads but allows phantom reads
Serializable	Locks entire table, prevents all anomalies	Uses SSI (no locks), prevents all anomalies
Key Takeaways
✅ PostgreSQL is safer at all isolation levels because:

It never allows dirty reads (even at READ UNCOMMITTED).
It uses Serializable Snapshot Isolation (SSI) instead of full table locks.
✅ MySQL needs table locks at SERIALIZABLE, which can hurt performance.
✅ If you want stronger consistency, PostgreSQL is better than MySQL at SERIALIZABLE.

VIPER: 
-> FIND load,unmarshal configfile
 JSON,TOML,YAML,ENV,INI
-> read config from environment variables or flag
   Override existing values,set default values
-> Read config from remote system
   ETCD,consul
-> Live watching and writng config file
   Reread changed file,save any modification
