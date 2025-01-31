createdb:
	docker exec -it 9b7f42121cbe createdb --username=postgres --owner=postgres simple_bank

dropdb:
	docker exec -it 9b7f42121cbe dropdb simple_bank

sqlc:
	sqlc generate	

.PHONY: createdb dropdb sqlc