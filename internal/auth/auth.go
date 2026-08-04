package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT payload used across the application: the standard
// registered claims (expiry, etc.) plus the authenticated user's ID.
type Claims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id"`
}

// GenerateToken creates and signs a JWT for the given user ID using HS256
// and the provided secret. The token expires 24 hours after issuance.
func GenerateToken(id int64, secret []byte) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			},
			UserID: id,
		},
	)
	return token.SignedString(secret)
}

// ParseToken validates the given JWT string against secret and returns its
// claims. It returns an error if the token is malformed, expired, or signed
// with an unexpected method or key.
func ParseToken(tokenString string, secret []byte) (*Claims, error) {
	claims := new(Claims)

	_, err := jwt.ParseWithClaims(
		tokenString, claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	return claims, nil
}
