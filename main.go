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

// @title							Simple Bank API
// @description						This is a simple bank API.
// @description					It provides APIs for the frontend to do following things:
// @description					1. Create and manage bank accounts, which are composed of owner’s name, balance, and currency.
// @description					2. Record all balance changes to each of the account. So every time some money is added to or subtracted from the account, an account entry record will be created.
// @description					3. Perform a money transfer between 2 accounts. This should happen within a transaction, so that either both accounts’ balance are updated successfully or none of them are.
// @description					Type "Bearer " followed by a space and then your token
// @contact.name				Javokhir Yulchiboev
// @contact.url					https://t.me/javakhiryu
// @contact.email				javakhiryulchibaev@gmail.com
// @host						localhost:8080
// @SecurityDefinitions.apiKey	Bearer
// @in							header
// @name						Authorization
// @version						1.0
// @BasePath					/
// @schemes						http
// @schemes						https
// @produce						json
// @consumes					json
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
