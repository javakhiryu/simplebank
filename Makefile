DB_URL=postgres://root:secret@localhost:5432/simplebank?sslmode=disable

postgres:
	 docker run --name postgres17 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:17-alpine
	    
createdb:
	docker exec -it postgres17 createdb --username=root --owner=root simplebank

dropdb:
	docker exec -it postgres17 dropdb simplebank

migrateinstall:
	$ curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.2/migrate.linux-amd64.tar.gz | tar xvz

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migrateuplast:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedownlast:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server: 
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go simplebank/db/sqlc Store 

db_docs:
	dbdocs build docs/db/db.dbml --password AksZX5wFaudwYhA

db_schema:
	dbml2sql --postgres -o docs/db/schema.sql docs/db/db.dbml

proto:
	rm -f pb/*.go
	rm -f docs/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=docs/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simplebank \
    proto/*.proto
	statik -src=./docs/swagger -dest=./docs

evans:
	evans --host localhost --port 9090  -r repl

swagger:
	swag init 

.PHONY: postgres createdb dropdb migrateup migratedown migrateuplast migratedownlast sqlc test migrateinstall server mock db_docs db_schema proto evans swagger