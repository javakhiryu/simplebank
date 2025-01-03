package db

//Convention: TestMain func is main entry point of all unit test inside one specific package

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:2694076@localhost:5434/simple_bank?sslmode=disable"
)

var testQueries *Queries

func TestMain(m *testing.M) {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to database", err)
	}
	testQueries = New(conn)

	os.Exit(m.Run())
}
