postgres:
	 docker run --name postgres17 -p 5434:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:17-alpine
	    
createdb:
	docker exec -it postgres17 createdb --username=root --owner=root simplebank

dropdb:
	docker exec -it postgres17 dropdb simplebank

migrateinstall:
	$ curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5434/simplebank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5434/simplebank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server: 
	go run main.go

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test migrateinstall server