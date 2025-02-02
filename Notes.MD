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