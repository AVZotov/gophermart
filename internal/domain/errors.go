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
	// ErrInvalidOrderID is returned when the provided order number fails Luhn validation.
	ErrInvalidOrderID = errors.New("invalid order id")
	// ErrOrderOwnedByAnotherUser is returned when the order number was already
	// uploaded by a different user, surfaced via the unique constraint on order_number.
	ErrOrderOwnedByAnotherUser = errors.New("order owned by another user")
	// ErrOrderAlreadyUploaded is returned when the same user re-uploads an order
	// number they already own; not a failure, the handler maps this to 200.
	ErrOrderAlreadyUploaded = errors.New("order already uploaded")
)
