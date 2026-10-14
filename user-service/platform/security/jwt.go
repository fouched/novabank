package security

import (
	"context"
	"time"

	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/golang-jwt/jwt/v5"
)

// Define a private unexported type for the context key to avoid collisions
type contextKey string

const claimsContextKey contextKey = "jwt_claims"

type Claims struct {
	UserID int64    `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"` // Array accommodates multi-role banking authorization maps
	jwt.RegisteredClaims
}

// GenerateToken acts precisely like the JwtService.generateToken logic in Spring
func GenerateToken(user *models.User, secret string, expiry time.Duration) (string, error) {
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Name
	}

	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "novabank.user-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GetClaims retrieves claims out of a native Go Request Context thread
func GetClaims(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}

// WithClaims injects decoded token contexts safely into an active Request lifecycle
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}
