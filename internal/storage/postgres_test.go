package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/AVZotov/gophermart/internal/domain"
	"github.com/AVZotov/gophermart/internal/storage"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupStorage(t *testing.T) *storage.PostgresStorage {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := postgrescontainer.Run(
		ctx,
		"postgres:16-alpine",
		postgrescontainer.WithDatabase("gophermart"),
		postgrescontainer.WithUsername("gophermart"),
		postgrescontainer.WithPassword("gophermart"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			require.NoError(t, container.Terminate(context.Background()))
		},
	)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	require.NoError(t, storage.RunMigrations(ctx, dsn))

	store, err := storage.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(store.Close)

	return store
}

func TestPostgresStorage_CreateUser_Success(t *testing.T) {
	store := setupStorage(t)
	ctx := context.Background()

	user := &domain.User{Login: "alice", PasswordHash: "hash"}
	id, err := store.CreateUser(ctx, user)

	require.NoError(t, err)
	require.Greater(t, id, int64(0))
}

func TestPostgresStorage_CreateUser_DuplicateLogin(t *testing.T) {
	store := setupStorage(t)
	ctx := context.Background()

	_, err := store.CreateUser(ctx, &domain.User{Login: "bob", PasswordHash: "hash1"})
	require.NoError(t, err)

	_, err = store.CreateUser(ctx, &domain.User{Login: "bob", PasswordHash: "hash2"})
	require.ErrorIs(t, err, domain.ErrUserExists)
}

func TestPostgresStorage_GetUserByLogin_NotFound(t *testing.T) {
	store := setupStorage(t)
	ctx := context.Background()

	user, err := store.GetUserByLogin(ctx, "ghost")

	require.ErrorIs(t, err, domain.ErrUserNotFound)
	require.Nil(t, user)
}

func TestPostgresStorage_GetUserByLogin_Success(t *testing.T) {
	store := setupStorage(t)
	ctx := context.Background()

	id, err := store.CreateUser(ctx, &domain.User{Login: "carol", PasswordHash: "hash"})
	require.NoError(t, err)

	user, err := store.GetUserByLogin(ctx, "carol")

	require.NoError(t, err)
	require.Equal(t, id, user.ID)
	require.Equal(t, "carol", user.Login)
	require.Equal(t, "hash", user.PasswordHash)
}
