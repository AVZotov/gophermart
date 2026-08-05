package storage

import (
	"context"

	"github.com/AVZotov/gophermart/internal/domain"
)

// Storage is the persistence interface required by the service layer. It
// performs CRUD operations only; business logic and validation live in
// internal/service.
type Storage interface {
	// CreateUser inserts user and returns its generated ID. It returns
	// domain.ErrUserExists if the login is already taken.
	CreateUser(ctx context.Context, user *domain.User) (int64, error)
	// GetUserByLogin fetches the user with the given login. It returns
	// domain.ErrUserNotFound if no such user exists.
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
	// CreateOrder inserts order for userID. It returns
	// domain.ErrOrderAlreadyUploaded if userID already uploaded this order
	// number, or domain.ErrOrderOwnedByAnotherUser if another user did.
	CreateOrder(ctx context.Context, order *domain.Order, userID int64) error
	// GetOrdersByUserID fetches all orders uploaded by userID, ordered by
	// upload time descending.
	GetOrdersByUserID(ctx context.Context, userID int64) ([]*domain.Order, error)
	// Close releases any resources held by the implementation.
	Close()
}
