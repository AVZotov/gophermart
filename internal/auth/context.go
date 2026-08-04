package auth

import "context"

type ctxKey int

const userIDKey ctxKey = iota

// WithUserID returns a copy of ctx carrying the authenticated user's ID.
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext retrieves the user ID stored by WithUserID. The second
// return value is false if ctx carries no user ID.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}
