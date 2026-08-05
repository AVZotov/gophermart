package service

import (
	"context"
	"errors"

	"github.com/AVZotov/gophermart/internal/auth"
	"github.com/AVZotov/gophermart/internal/domain"
	"github.com/AVZotov/gophermart/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

// Service implements the application's business logic, delegating
// persistence to the injected storage.Storage.
type Service struct {
	store  storage.Storage
	secret []byte
}

// NewService constructs a Service backed by store for persistence and
// secret for signing issued JWTs.
func NewService(store storage.Storage, secret []byte) *Service {
	return &Service{
		store:  store,
		secret: secret,
	}
}

// Register hashes rawPassword, creates a new user with the given login, and
// returns a signed JWT for the new user. It returns domain.ErrUserExists if
// login is already taken.
func (s *Service) Register(ctx context.Context, login, rawPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", domain.ErrPasswordTooLong
		}
		return "", err
	}
	user := domain.User{
		Login:        login,
		PasswordHash: string(hash),
	}
	id, err := s.store.CreateUser(ctx, &user)
	if err != nil {
		return "", err
	}
	token, err := auth.GenerateToken(id, s.secret)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Login verifies login and rawPassword against the stored user and, on
// success, returns a signed JWT. It returns domain.ErrInvalidCredentials if
// the login does not exist or the password does not match; the two cases
// are not distinguished, to avoid leaking which logins are registered.
func (s *Service) Login(ctx context.Context, login, rawPassword string) (string, error) {
	user, err := s.store.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(rawPassword)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	return auth.GenerateToken(user.ID, s.secret)
}

// UploadOrder validates orderNumber (Luhn) and registers it for userID. It
// returns domain.ErrInvalidOrderID if the number fails validation,
// domain.ErrOrderAlreadyUploaded if userID already uploaded this order, or
// domain.ErrOrderOwnedByAnotherUser if another user did.
func (s *Service) UploadOrder(ctx context.Context, userID int64, orderNumber string) error {
	order, err := domain.NewOrder(orderNumber)
	if err != nil {
		return err
	}

	return s.store.CreateOrder(ctx, order, userID)
}

// GetOrders returns all orders uploaded by userID, ordered by upload time
// descending.
func (s *Service) GetOrders(ctx context.Context, userID int64) ([]*domain.Order, error) {
	return s.store.GetOrdersByUserID(ctx, userID)
}
