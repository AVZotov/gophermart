package storage_test

import (
	"context"
	"errors"
	"sync"
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

func TestPostgresStorage_CreateOrder_Success(t *testing.T) {
	store := setupStorage(t)
	ctx := context.Background()

	userID, err := store.CreateUser(ctx, &domain.User{Login: "dave", PasswordHash: "hash"})
	require.NoError(t, err)

	order, err := domain.NewOrder("79927398713")
	require.NoError(t, err)

	err = store.CreateOrder(ctx, order, userID)

	require.NoError(t, err)
}

func TestPostgresStorage_CreateOrder_AlreadyUploadedBySameUser(t *testing.T) {
	store := setupStorage(t)
	ctx := context.Background()

	userID, err := store.CreateUser(ctx, &domain.User{Login: "erin", PasswordHash: "hash"})
	require.NoError(t, err)

	order, err := domain.NewOrder("79927398713")
	require.NoError(t, err)

	require.NoError(t, store.CreateOrder(ctx, order, userID))

	err = store.CreateOrder(ctx, order, userID)

	require.ErrorIs(t, err, domain.ErrOrderAlreadyUploaded)
}

func TestPostgresStorage_CreateOrder_OwnedByAnotherUser(t *testing.T) {
	store := setupStorage(t)
	ctx := context.Background()

	userID1, err := store.CreateUser(ctx, &domain.User{Login: "frank", PasswordHash: "hash"})
	require.NoError(t, err)
	userID2, err := store.CreateUser(ctx, &domain.User{Login: "grace", PasswordHash: "hash"})
	require.NoError(t, err)

	order, err := domain.NewOrder("79927398713")
	require.NoError(t, err)

	require.NoError(t, store.CreateOrder(ctx, order, userID1))

	err = store.CreateOrder(ctx, order, userID2)

	require.ErrorIs(t, err, domain.ErrOrderOwnedByAnotherUser)
}

func TestPostgresStorage_GetOrdersByUserID_Success(t *testing.T) {
	store := setupStorage(t)
	ctx := context.Background()

	userID, err := store.CreateUser(ctx, &domain.User{Login: "heidi", PasswordHash: "hash"})
	require.NoError(t, err)

	order1, err := domain.NewOrder("79927398713")
	require.NoError(t, err)
	order2, err := domain.NewOrder("12345678903")
	require.NoError(t, err)

	require.NoError(t, store.CreateOrder(ctx, order1, userID))
	require.NoError(t, store.CreateOrder(ctx, order2, userID))

	orders, err := store.GetOrdersByUserID(ctx, userID)

	require.NoError(t, err)
	require.Len(t, orders, 2)
	numbers := []string{orders[0].Number, orders[1].Number}
	require.ElementsMatch(t, []string{"79927398713", "12345678903"}, numbers)
	require.Equal(t, domain.OrderStatusNew, orders[0].Status)
}

func TestPostgresStorage_GetOrdersByUserID_Empty(t *testing.T) {
	store := setupStorage(t)
	ctx := context.Background()

	userID, err := store.CreateUser(ctx, &domain.User{Login: "ivan", PasswordHash: "hash"})
	require.NoError(t, err)

	orders, err := store.GetOrdersByUserID(ctx, userID)

	require.NoError(t, err)
	require.Empty(t, orders)
}

// TestPostgresStorage_CreateOrder_ConcurrentRace uploads the same order
// number for two different users concurrently, verifying that the unique
// constraint on order_number (not an in-Go check-then-write) is what
// arbitrates the race: exactly one caller must succeed and the other must
// observe domain.ErrOrderOwnedByAnotherUser. Run with -race to also confirm
// there is no data race in the storage implementation itself.
func TestPostgresStorage_CreateOrder_ConcurrentRace(t *testing.T) {
	store := setupStorage(t)
	ctx := context.Background()

	userID1, err := store.CreateUser(ctx, &domain.User{Login: "judy", PasswordHash: "hash"})
	require.NoError(t, err)
	userID2, err := store.CreateUser(ctx, &domain.User{Login: "karl", PasswordHash: "hash"})
	require.NoError(t, err)

	order, err := domain.NewOrder("79927398713")
	require.NoError(t, err)

	var wg sync.WaitGroup
	errs := make([]error, 2)
	userIDs := []int64{userID1, userID2}

	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = store.CreateOrder(ctx, order, userIDs[i])
		}(i)
	}
	wg.Wait()

	successes, conflicts := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrOrderOwnedByAnotherUser):
			conflicts++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}

	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)
}
