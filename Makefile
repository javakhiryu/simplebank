postgres:
	 docker run --name postgres17 -p 5434:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=2694076 -d postgres:17-alpine
	    
createdb:
	docker exec -it postgres17 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres17 dropdb simple_bank	

migrateup:
	migrate -path db/migration -database "postgresql://root:2694076@localhost:5434/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:2694076@localhost:5434/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test