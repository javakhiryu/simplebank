package db

import (
	"context"
	"database/sql"
	"simplebank/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashedPassword(util.RandomString(6))
	require.NoError(t, err)
	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomString(6),
		Email:          util.RandomString(6) + "@gmail.com",
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}
func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestUpdateUserHashedPassword(t *testing.T) {
	newHashedPassword, err := util.HashedPassword(util.RandomString(6))
	require.NoError(t, err)
	user1 := createRandomUser(t)

	arg := UpdateUserHashedPasswordParams{
		Username:       user1.Username,
		HashedPassword: newHashedPassword,
	}
	user2, err := testQueries.UpdateUserHashedPassword(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.NotEqual(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, arg.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.WithinDuration(t, time.Now(), user2.PasswordChangedAt, time.Second)
}

func TestUpdateUsersFullname(t *testing.T) {
	oldUser := createRandomUser(t)
	newFullName := util.RandomString(6)
	arg := UpdateUserParams{
		Username: oldUser.Username,
		FullName: sql.NullString{
			String: newFullName,
			Valid:  true,
		},
	}

	newUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEqual(t, oldUser.FullName, newUser.FullName)
	require.Equal(t, newFullName, newUser.FullName)
}

func TestUpdateUsersEmail(t *testing.T) {
	oldUser := createRandomUser(t)
	newEmail := util.RandomString(6) + "@gmail.com"
	arg := UpdateUserParams{
		Username: oldUser.Username,
		Email: sql.NullString{
			String: newEmail,
			Valid:  true,
		},
	}

	newUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEqual(t, oldUser.Email, newUser.Email)
	require.Equal(t, newEmail, newUser.Email)
}

func TestUpdateUsersFullnameEmail(t *testing.T) {
	oldUser := createRandomUser(t)
	newFullName := util.RandomString(6)
	newEmail := util.RandomString(6) + "@gmail.com"
	arg := UpdateUserParams{
		Username: oldUser.Username,
		FullName: sql.NullString{
			String: newFullName,
			Valid:  true,
		},
		Email: sql.NullString{
			String: newEmail,
			Valid:  true,
		},
	}

	newUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEqual(t, oldUser.FullName, newUser.FullName)
	require.Equal(t, newFullName, newUser.FullName)
	require.NotEqual(t, oldUser.Email, newUser.Email)
	require.Equal(t, newEmail, newUser.Email)
}
