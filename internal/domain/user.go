package domain

// User represents a registered account, identified by a unique login and
// authenticated via a bcrypt password hash.
type User struct {
	ID           int64
	Login        string
	PasswordHash string
}
