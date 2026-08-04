package auth

import (
	"net/http"
	"strings"
)

// AuthMiddleware returns HTTP middleware that requires a valid
// "Authorization: Bearer <token>" header signed with secret. It responds
// with 401 Unauthorized if the header is missing or the token fails to
// parse; otherwise it injects the token's user ID into the request context
// via WithUserID before calling the next handler.
func AuthMiddleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				h := r.Header.Get("Authorization")
				tokenString, ok := strings.CutPrefix(h, "Bearer ")
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				claims, err := ParseToken(tokenString, secret)
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				ctx := WithUserID(r.Context(), claims.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}
