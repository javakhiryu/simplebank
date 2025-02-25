package main

import (
	"database/sql"
	"log"
	"simplebank/api"
	db "simplebank/db/sqlc"
	"simplebank/util"

	_ "simplebank/docs"

	_ "github.com/lib/pq"
)

//	@title			Simple Bank API
//	@version		1.0
//	@description	This is a simple bank API
//	@host			api.javakhiryu-simplebank.click
//	@BasePath
//	@schemes	http
//	@schemes	https
//	@produce	json
//	@consumes	json
func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to database", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(store, config)
	if err != nil {
		log.Fatal("Cannot create server:", err)
		return
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("Cannot start the server:", err)
	}

}
