package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"simplebank/api"
	db "simplebank/db/sqlc"
	"simplebank/gapi"
	"simplebank/pb"
	"simplebank/util"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "simplebank/docs"
	_ "simplebank/docs/statik"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

// @title			Simple Bank API
// @description	This is a simple bank API.
// @description	It provides APIs for the frontend to do following things:
// @description	1. Create and manage bank accounts, which are composed of owner’s name, balance, and currency.
// @description	2. Record all balance changes to each of the account. So every time some money is added to or subtracted from the account, an account entry record will be created.
// @description	3. Perform a money transfer between 2 accounts. This should happen within a transaction, so that either both accounts’ balance are updated successfully or none of them are.
// @description
// @description	Feel free to contact me if you have any questions
// @description
// @description				GitHub Repository:
// @contact.name				Javakhir Yu
// @contact.url				https://github.com/javakhiryu/simplebank
// @contact.email				javakhiryulchibaev@gmail.com
// @host						localhost:8080
// @SecurityDefinitions.apiKey	Bearer
// @in							header
// @name						Authorization
// @description				Type "Bearer " followed by a space and then your token
// @version					1.0
// @BasePath					/
// @schemes					http
// @schemes					https
// @produce					json
// @consumes				json
func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to database")
	}

	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)
	go runGatewayServer(config, store)
	runGRPCServer(config, store)

}

func runDBMigration(configURL string, dbSource string) {
	migration, err := migrate.New(configURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance:")
	}
	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msg("failed to run migration up:")
	}
	log.Info().Msg("db migrated successfully")
}

func runGRPCServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(store, config)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create server")
		return
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}
	log.Info().Msgf("start gRPC server at %s", listener.Addr().String())

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start grpc server")
	}
}

func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(store, config)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create server")
		return
	}
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot register handler server")
		return
	}
	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create statik fs")
	}
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}
	log.Info().Msgf("start HTTP gateway server at %s", listener.Addr().String())
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start http server")
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(store, config)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create server")
		return
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot start the server")
	}
}
