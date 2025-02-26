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

// @title						Simple Bank API
// @description					This is a simple bank API
// @host						api.javakhiryu-simplebank.click
// @SecurityDefinitions.apiKey	Bearer
// @in							header
// @name						Authorization
// @description					Type "Bearer " followed by a space and then your token
// @contact.name				Javokhir Yulchiboev
// @contact.url					https://t.me/javakhiryu
// @contact.email				javakhiryulchibaev@gmail.com
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
