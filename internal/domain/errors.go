package domain

import "errors"

var (
	// ErrUserExists is returned when registration is attempted with a login
	// that already exists, typically surfaced via a unique constraint violation.
	ErrUserExists = errors.New("user already exists")
	// ErrUserNotFound is returned when no user matches the requested login.
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidCredentials is returned on login when the login does not exist
	// or the provided password does not match the stored hash. The two cases
	// are deliberately not distinguished to avoid leaking which logins exist.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrPasswordTooLong is returned when the provided password exceeds
	// bcrypt's 72-byte input limit.
	ErrPasswordTooLong = errors.New("password too long")
)
