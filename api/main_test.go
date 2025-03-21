package api

//Convention: TestMain func is main entry point of all unit test inside one specific package

import (
	"os"
	db "simplebank/db/sqlc"
	"simplebank/util"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// NewTestServer creates a new instance of Server for testing purposes.
// It takes a testing object and a db.Store as parameters, configures a
// util.Config with random TokenSymetricKey and AccessTokenDuration set to one minute,
// and returns the initialized Server. It requires no error from NewServer.

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymetricKey:     util.RandomString(32),
		AccessTokenDuration:  time.Minute,
		RefreshTokenDuration: time.Hour,
	}
	server, err := NewServer(store, config)
	require.NoError(t, err)
	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
