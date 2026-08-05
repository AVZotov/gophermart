package service

import (
	"context"
	"testing"

	"github.com/AVZotov/gophermart/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type fakeStorage struct {
	createUserFunc        func(ctx context.Context, user *domain.User) (int64, error)
	getUserByLoginFunc    func(ctx context.Context, login string) (*domain.User, error)
	createOrderFunc       func(ctx context.Context, order *domain.Order, userID int64) error
	getOrdersByUserIDFunc func(ctx context.Context, userID int64) ([]*domain.Order, error)
}

func (f *fakeStorage) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	return f.createUserFunc(ctx, user)
}

func (f *fakeStorage) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	return f.getUserByLoginFunc(ctx, login)
}

func (f *fakeStorage) CreateOrder(ctx context.Context, order *domain.Order, userID int64) error {
	return f.createOrderFunc(ctx, order, userID)
}

func (f *fakeStorage) GetOrdersByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	return f.getOrdersByUserIDFunc(ctx, userID)
}

func (f *fakeStorage) Close() {}

var testSecret = []byte("test-secret")

func TestService_Register_Success(t *testing.T) {
	store := &fakeStorage{
		createUserFunc: func(ctx context.Context, user *domain.User) (int64, error) {
			assert.Equal(t, "alice", user.Login)
			assert.NotEmpty(t, user.PasswordHash)
			return 1, nil
		},
	}
	svc := NewService(store, testSecret)

	token, err := svc.Register(context.Background(), "alice", "password123")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestService_Register_DuplicateLogin(t *testing.T) {
	store := &fakeStorage{
		createUserFunc: func(ctx context.Context, user *domain.User) (int64, error) {
			return 0, domain.ErrUserExists
		},
	}
	svc := NewService(store, testSecret)

	token, err := svc.Register(context.Background(), "alice", "password123")

	require.ErrorIs(t, err, domain.ErrUserExists)
	assert.Empty(t, token)
}

func TestService_Login_Success(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	store := &fakeStorage{
		getUserByLoginFunc: func(ctx context.Context, login string) (*domain.User, error) {
			assert.Equal(t, "alice", login)
			return &domain.User{ID: 1, Login: "alice", PasswordHash: string(hash)}, nil
		},
	}
	svc := NewService(store, testSecret)

	token, err := svc.Login(context.Background(), "alice", "password123")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestService_Login_WrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	store := &fakeStorage{
		getUserByLoginFunc: func(ctx context.Context, login string) (*domain.User, error) {
			return &domain.User{ID: 1, Login: "alice", PasswordHash: string(hash)}, nil
		},
	}
	svc := NewService(store, testSecret)

	token, err := svc.Login(context.Background(), "alice", "wrong-password")

	require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	assert.Empty(t, token)
}

func TestService_Login_UserNotFound(t *testing.T) {
	store := &fakeStorage{
		getUserByLoginFunc: func(ctx context.Context, login string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	svc := NewService(store, testSecret)

	token, err := svc.Login(context.Background(), "ghost", "password123")

	require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	assert.Empty(t, token)
}

func TestService_UploadOrder_Success(t *testing.T) {
	store := &fakeStorage{
		createOrderFunc: func(ctx context.Context, order *domain.Order, userID int64) error {
			assert.Equal(t, "79927398713", order.Number)
			assert.Equal(t, int64(1), userID)
			return nil
		},
	}
	svc := NewService(store, testSecret)

	err := svc.UploadOrder(context.Background(), 1, "79927398713")

	require.NoError(t, err)
}

func TestService_UploadOrder_InvalidNumber(t *testing.T) {
	store := &fakeStorage{
		createOrderFunc: func(ctx context.Context, order *domain.Order, userID int64) error {
			t.Fatal("storage should not be called for an invalid order number")
			return nil
		},
	}
	svc := NewService(store, testSecret)

	err := svc.UploadOrder(context.Background(), 1, "12345678900")

	require.ErrorIs(t, err, domain.ErrInvalidOrderID)
}

func TestService_UploadOrder_AlreadyUploadedByOwner(t *testing.T) {
	store := &fakeStorage{
		createOrderFunc: func(ctx context.Context, order *domain.Order, userID int64) error {
			return domain.ErrOrderAlreadyUploaded
		},
	}
	svc := NewService(store, testSecret)

	err := svc.UploadOrder(context.Background(), 1, "79927398713")

	require.ErrorIs(t, err, domain.ErrOrderAlreadyUploaded)
}

func TestService_UploadOrder_OwnedByAnotherUser(t *testing.T) {
	store := &fakeStorage{
		createOrderFunc: func(ctx context.Context, order *domain.Order, userID int64) error {
			return domain.ErrOrderOwnedByAnotherUser
		},
	}
	svc := NewService(store, testSecret)

	err := svc.UploadOrder(context.Background(), 1, "79927398713")

	require.ErrorIs(t, err, domain.ErrOrderOwnedByAnotherUser)
}

func TestService_GetOrders_Success(t *testing.T) {
	want := []*domain.Order{
		{Number: "79927398713", Status: domain.OrderStatusNew},
	}
	store := &fakeStorage{
		getOrdersByUserIDFunc: func(ctx context.Context, userID int64) ([]*domain.Order, error) {
			assert.Equal(t, int64(1), userID)
			return want, nil
		},
	}
	svc := NewService(store, testSecret)

	orders, err := svc.GetOrders(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, want, orders)
}

func TestService_GetOrders_StorageError(t *testing.T) {
	store := &fakeStorage{
		getOrdersByUserIDFunc: func(ctx context.Context, userID int64) ([]*domain.Order, error) {
			return nil, assert.AnError
		},
	}
	svc := NewService(store, testSecret)

	orders, err := svc.GetOrders(context.Background(), 1)

	require.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, orders)
}
