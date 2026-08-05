package storage

import (
	"context"
	"errors"
	"fmt"

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

// CreateOrder inserts order for userID, relying on the unique constraint on
// order_number to detect duplicates atomically (no separate existence
// check, avoiding a TOCTOU race). On a unique violation (pgerrcode 23505)
// it looks up the existing owner to distinguish domain.ErrOrderAlreadyUploaded
// (same user) from domain.ErrOrderOwnedByAnotherUser (different user).
func (s *PostgresStorage) CreateOrder(ctx context.Context, order *domain.Order, userID int64) error {
	insertQuery := `INSERT INTO orders (order_number, user_id) VALUES ($1, $2)`

	_, err := s.pool.Exec(ctx, insertQuery, order.Number, userID)
	if err == nil {
		return nil
	}

	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if !ok || pgErr.Code != "23505" {
		return err
	}

	selectQuery := `SELECT user_id FROM orders WHERE order_number = $1`
	row := s.pool.QueryRow(ctx, selectQuery, order.Number)

	var existingUserID int64
	if err := row.Scan(&existingUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("order %q vanished after unique violation: %w", order.Number, err)
		}
		return err
	}

	if existingUserID != userID {
		return domain.ErrOrderOwnedByAnotherUser
	}
	return domain.ErrOrderAlreadyUploaded
}

// GetOrdersByUserID fetches all orders uploaded by userID, ordered by
// upload time descending. It returns a nil slice (not an error) if the
// user has no orders.
func (s *PostgresStorage) GetOrdersByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	query := `SELECT order_number, status, accrual_cents, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.Number, &o.Status, &o.AccrualCents, &o.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
