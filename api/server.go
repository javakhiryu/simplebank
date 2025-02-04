package api

import (
	"fmt"
	db "simplebank/db/sqlc"

	"simplebank/token"
	"simplebank/util"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(store db.Store, config util.Config) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}
	
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}
	
	server.setupRounter()

	return server, nil
}

func (server *Server) setupRounter(){
	router := gin.Default()

	router.POST("/createUser", server.createUser)
	router.POST("/login", server.loginUser)
	router.POST("/account", server.createAccount)

	router.GET("/account/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)
	router.GET("/getUser/:username", server.getUser)

	router.PATCH("/updatePassword", server.updateUserHashedPassword)

	router.POST("/createTransfer", server.createTransfer)

	router.DELETE("/account/:id", server.deleteAccount)
	
	

	server.router = router

}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
