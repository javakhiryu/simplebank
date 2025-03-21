package gapi

import (
	"context"
	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	token "simplebank/token"
	"simplebank/util"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateUserAPi(t *testing.T) {
	password := util.RandomString(6)
	hashedPassword, err := util.HashedPassword(password)
	if err != nil {
		return
	}
	user := randomUser(hashedPassword)
	newName := util.RandomOwner()
	newEmail := newName + "@mail.com"
	invalidEmailAddress := "invalid email"

	testCases := []struct {
		name          string
		req           *pb.UpdateUserRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.UpdateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: pgtype.Text{
						String: newName,
						Valid:  true,
					},
					Email: pgtype.Text{
						String: newEmail,
						Valid:  true,
					},
				}

				store.EXPECT().UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(
						db.User{
							Username:          user.Username,
							FullName:          newName,
							Email:             newEmail,
							PasswordChangedAt: user.PasswordChangedAt,
							CreatedAt:         user.CreatedAt,
							IsEmailVerified:   user.IsEmailVerified,
						}, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAuth(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				updatedUser := res.GetUser()
				require.Equal(t, user.Username, updatedUser.Username)
				require.Equal(t, newName, updatedUser.FullName)
				require.Equal(t, newEmail, updatedUser.Email)
			},
		},
		{
			name: "UserNotFound",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: pgtype.Text{
						String: newName,
						Valid:  true,
					},
					Email: pgtype.Text{
						String: newEmail,
						Valid:  true,
					},
				}

				store.EXPECT().UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(
						db.User{}, db.ErrNoRowsFound)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAuth(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())

			},
		},
		{
			name: "ExpiredToken",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: pgtype.Text{
						String: newName,
						Valid:  true,
					},
					Email: pgtype.Text{
						String: newEmail,
						Valid:  true,
					},
				}

				store.EXPECT().UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAuth(t, tokenMaker, user.Username, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())

			},
		},
		{
			name: "NoAuthorization",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: pgtype.Text{
						String: newName,
						Valid:  true,
					},
					Email: pgtype.Text{
						String: newEmail,
						Valid:  true,
					},
				}

				store.EXPECT().UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())

			},
		},
		{
			name: "InvalidEmailAddress",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &invalidEmailAddress,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: pgtype.Text{
						String: newName,
						Valid:  true,
					},
					Email: pgtype.Text{
						String: newEmail,
						Valid:  true,
					},
				}

				store.EXPECT().UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAuth(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())

			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			store := mockdb.NewMockStore(mockCtrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, nil)

			ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.UpdateUser(ctx, tc.req)

			tc.checkResponse(t, res, err)
		})

	}
}
