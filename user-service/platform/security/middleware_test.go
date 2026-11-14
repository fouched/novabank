package security

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecret = "our-super-secure-dev-only-jwt-secret-key-32-chars"
	testIssuer = "novabank.user-service"
)

// helper to mint custom tokens for edge-case testing
func createTestToken(t *testing.T, claims Claims, method jwt.SigningMethod, key any) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return signed
}

func TestJWTMiddleware(t *testing.T) {
	// Generate an RSA key once to test algorithm substitution rejection
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	validClaims := Claims{
		UserID: 42,
		Email:  "jane.doe@novabank.co.za",
		Roles:  []models.RoleName{models.RoleCustomer},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    testIssuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	tests := []struct {
		name           string
		authHeader     string
		wantStatus     int
		wantErrorMsg   string
		wantClaimEmail string
	}{
		{
			name:           "valid token injects claims into context",
			authHeader:     "Bearer " + createTestToken(t, validClaims, jwt.SigningMethodHS256, []byte(testSecret)),
			wantStatus:     http.StatusOK,
			wantClaimEmail: "jane.doe@novabank.co.za",
		},
		{
			name:         "missing authorization header returns 401",
			authHeader:   "",
			wantStatus:   http.StatusUnauthorized,
			wantErrorMsg: "missing or invalid authorization header",
		},
		{
			name:         "malformed prefix returns 401",
			authHeader:   "Token abc.def.ghi",
			wantStatus:   http.StatusUnauthorized,
			wantErrorMsg: "missing or invalid authorization header",
		},
		{
			name:         "empty bearer token returns 401",
			authHeader:   "Bearer ",
			wantStatus:   http.StatusUnauthorized,
			wantErrorMsg: "missing or invalid authorization header",
		},
		{
			name: "expired token returns 401",
			authHeader: "Bearer " + createTestToken(t, Claims{
				UserID: 42,
				Email:  "jane.doe@novabank.co.za",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    testIssuer,
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(-5 * time.Minute)),
				},
			}, jwt.SigningMethodHS256, []byte(testSecret)),
			wantStatus:   http.StatusUnauthorized,
			wantErrorMsg: "invalid or expired authentication token",
		},
		{
			name: "wrong issuer returns 401",
			authHeader: "Bearer " + createTestToken(t, Claims{
				UserID: 42,
				Email:  "jane.doe@novabank.co.za",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "malicious-service",
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
				},
			}, jwt.SigningMethodHS256, []byte(testSecret)),
			wantStatus:   http.StatusUnauthorized,
			wantErrorMsg: "invalid or expired authentication token",
		},
		{
			name:         "wrong signing secret returns 401",
			authHeader:   "Bearer " + createTestToken(t, validClaims, jwt.SigningMethodHS256, []byte("wrong-secret")),
			wantStatus:   http.StatusUnauthorized,
			wantErrorMsg: "invalid or expired authentication token",
		},
		{
			name:         "non-HMAC algorithm (RS256) is rejected with 401",
			authHeader:   "Bearer " + createTestToken(t, validClaims, jwt.SigningMethodRS256, rsaKey),
			wantStatus:   http.StatusUnauthorized,
			wantErrorMsg: "invalid or expired authentication token",
		},
	}

	// Dummy downstream handler verifying context propagation
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r.Context())
		if !ok {
			t.Error("expected claims in context, but none found")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(claims.Email))
	})

	protected := JWTMiddleware(testSecret, testIssuer)(nextHandler)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/users/profile", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("got status %d, want %d (body: %s)", rec.Code, tc.wantStatus, rec.Body.String())
			}

			if tc.wantStatus == http.StatusOK {
				if got := rec.Body.String(); got != tc.wantClaimEmail {
					t.Errorf("got email %q, want %q", got, tc.wantClaimEmail)
				}
				return
			}

			// Verify JSON Content-Type and error payload on 401 responses
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("got Content-Type %q, want %q", ct, "application/json")
			}

			var errResp map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
				t.Fatalf("failed to unmarshal error JSON: %v", err)
			}
			if errResp["error"] != tc.wantErrorMsg {
				t.Errorf("got error %q, want %q", errResp["error"], tc.wantErrorMsg)
			}
		})
	}
}

func TestRequireAnyRole(t *testing.T) {
	tests := []struct {
		name         string
		injectClaims *Claims
		allowedRoles []models.RoleName
		wantStatus   int
	}{
		{
			name:         "allows request when user holds one of the allowed roles",
			injectClaims: &Claims{UserID: 1, Roles: []models.RoleName{models.RoleAdmin, models.RoleCustomer}},
			allowedRoles: []models.RoleName{models.RoleAdmin, models.RoleCustomer},
			wantStatus:   http.StatusOK,
		},
		{
			name:         "forbids request when user lacks allowed roles",
			injectClaims: &Claims{UserID: 2, Roles: []models.RoleName{models.RoleCustomer}},
			allowedRoles: []models.RoleName{models.RoleAdmin},
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "returns 401 when claims are missing from context",
			injectClaims: nil,
			allowedRoles: []models.RoleName{models.RoleAdmin},
			wantStatus:   http.StatusUnauthorized,
		},
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/users/all", nil)
			if tc.injectClaims != nil {
				req = req.WithContext(WithClaims(req.Context(), tc.injectClaims))
			}

			rec := httptest.NewRecorder()
			handler := RequireAnyRole(tc.allowedRoles...)(nextHandler)
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}
