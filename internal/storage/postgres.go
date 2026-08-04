package storage

import (
	"context"
	"errors"

	"github.com/AVZotov/gophermart/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStorage is a pgx/v5-backed implementation of Storage.
type PostgresStorage struct {
	pool *pgxpool.Pool
}

var _ Storage = (*PostgresStorage)(nil)

// New creates a PostgresStorage connected to dsn, verifying connectivity
// with a ping before returning.
func New(ctx context.Context, dsn string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStorage{pool: pool}, nil
}

// Close closes the underlying connection pool.
func (s *PostgresStorage) Close() {
	s.pool.Close()
}

// CreateUser inserts user (login and password hash) and returns its
// generated ID. It returns domain.ErrUserExists if the login already exists
// (unique constraint violation, pgerrcode 23505).
func (s *PostgresStorage) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`
	row := s.pool.QueryRow(ctx, query, user.Login, user.PasswordHash)

	var id int64
	if err := row.Scan(&id); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return 0, domain.ErrUserExists
		}
		return 0, err
	}
	return id, nil
}

// GetUserByLogin fetches the user with the given login. It returns
// domain.ErrUserNotFound if no such user exists.
func (s *PostgresStorage) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	query := `SELECT id, login, password_hash FROM users WHERE login = $1`
	row := s.pool.QueryRow(ctx, query, login)

	var user domain.User
	if err := row.Scan(&user.ID, &user.Login, &user.PasswordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
