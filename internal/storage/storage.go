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
	// Close releases any resources held by the implementation.
	Close()
}
