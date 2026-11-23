package security

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware validates bearer credentials and binds identity states to the execution Context
func JWTMiddleware(secret, expectedIssuer string) func(http.Handler) http.Handler {
	secretBytes := []byte(secret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				writeJSONError(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(
				tokenStr,
				claims,
				func(t *jwt.Token) (any, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
					}
					return secretBytes, nil
				},
				jwt.WithIssuer(expectedIssuer),
			)

			if err != nil || !token.Valid {
				writeJSONError(w, http.StatusUnauthorized, "invalid or expired authentication token")
				return
			}

			// Explicit over implicit: Bind claims metadata value right to the target Request Context
			ctx := WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAnyPermission blocks execution paths if the caller lacks authorisation matching parameters
func RequireAnyPermission(registry *PermissionRegistry, permissions ...models.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r.Context())
			if !ok {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			if registry.HasAnyPermission(claims.Roles, permissions...) {
				next.ServeHTTP(w, r)
				return
			}

			writeJSONError(w, http.StatusForbidden, "forbidden: insufficient permissions")
			return
		})
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `{"error":%q}`, message)
}
