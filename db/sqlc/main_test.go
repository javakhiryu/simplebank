package db

//Convention: TestMain func is main entry point of all unit test inside one specific package

import (
	"context"
	"log"
	"os"
	"simplebank/util"
	"testing"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testStore Store

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to database", err)
	}

	testStore = NewStore(conn)

	os.Exit(m.Run())
}
