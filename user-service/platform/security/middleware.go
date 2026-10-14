package security

import (
	"net/http"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware validates bearer credentials and binds identity states to the execution Context
func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired authentication token"}`, http.StatusUnauthorized)
				return
			}

			// Explicit over implicit: Bind claims metadata value right to the target Request Context
			ctx := WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAnyRole blocks execution paths if the caller lacks authorization matching parameters
func RequireAnyRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r.Context())
			if !ok {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Validate if user has at least one matching allowed role
			hasRole := false
			for _, userRole := range claims.Roles {
				if slices.Contains(allowedRoles, userRole) {
					hasRole = true
				}

				if hasRole {
					break
				}
			}

			if !hasRole {
				http.Error(w, `{"error":"forbidden: insufficient permission spatial profile"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
