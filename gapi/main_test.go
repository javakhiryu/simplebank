package gapi

import (
	"context"
	"fmt"
	db "simplebank/db/sqlc"
	"simplebank/token"
	"simplebank/util"
	"simplebank/worker"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func newTestServer(t *testing.T, store db.Store, taskDistibutor worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymetricKey:     util.RandomString(32),
		AccessTokenDuration:  time.Minute,
		RefreshTokenDuration: time.Hour,
	}
	server, err := NewServer(store, config, taskDistibutor)
	require.NoError(t, err)
	return server
}

func newContextWithAuth(t *testing.T, tokenMaker token.Maker, username string, tokenDuration time.Duration) context.Context {
	ctx := context.Background()
	accessToken, _, err := tokenMaker.CreateToken(username, tokenDuration)
	require.NoError(t, err)
	md := metadata.MD{
		authorizationHeaderKey: []string{
			fmt.Sprintf("%s %s", authorizationBearer, accessToken),
		},
	}
	return metadata.NewIncomingContext(ctx, md)
}
