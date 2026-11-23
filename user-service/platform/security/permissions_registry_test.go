package security

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fouched/novabank/users/internal/core/models"
)

type mockPermissionRepo struct {
	data map[models.RoleName]map[models.Permission]struct{}
	err  error
}

func (m *mockPermissionRepo) LoadRolePermissions(ctx context.Context) (map[models.RoleName]map[models.Permission]struct{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

func newTestRegistry(t *testing.T) *PermissionRegistry {
	t.Helper()
	repo := &mockPermissionRepo{
		data: map[models.RoleName]map[models.Permission]struct{}{
			models.RoleCustomer: {
				models.UserReadOwn:    {},
				models.AccountReadOwn: {},
			},
			models.RoleAdmin: {
				models.UserReadAll:   {},
				models.UserUpdate:    {},
				models.AccountSearch: {},
			},
		},
	}

	registry := NewPermissionRegistry(repo)
	if err := registry.Reload(context.Background()); err != nil {
		t.Fatalf("failed to load test registry: %v", err)
	}
	return registry
}

func TestPermissionRegistry_HasAnyPermission(t *testing.T) {
	registry := newTestRegistry(t)

	tests := []struct {
		name        string
		roles       []models.RoleName
		permissions []models.Permission
		want        bool
	}{
		{
			name:        "returns true when single role has matching permission",
			roles:       []models.RoleName{models.RoleCustomer},
			permissions: []models.Permission{models.UserReadOwn},
			want:        true,
		},
		{
			name:        "returns true when one of multiple required permissions matches",
			roles:       []models.RoleName{models.RoleCustomer},
			permissions: []models.Permission{models.UserReadAll, models.AccountReadOwn},
			want:        true,
		},
		{
			name:        "returns true when second role holds the required permission",
			roles:       []models.RoleName{models.RoleCustomer, models.RoleAdmin},
			permissions: []models.Permission{models.UserUpdate},
			want:        true,
		},
		{
			name:        "returns false when role lacks required permission",
			roles:       []models.RoleName{models.RoleCustomer},
			permissions: []models.Permission{models.UserUpdate},
			want:        false,
		},
		{
			name:        "returns false for unknown role",
			roles:       []models.RoleName{"UNKNOWN_ROLE"},
			permissions: []models.Permission{models.UserReadOwn},
			want:        false,
		},
		{
			name:        "returns false when user has no roles",
			roles:       nil,
			permissions: []models.Permission{models.UserReadOwn},
			want:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := registry.HasAnyPermission(tc.roles, tc.permissions...)
			if got != tc.want {
				t.Errorf("HasAnyPermission(%v, %v) = %v, want %v", tc.roles, tc.permissions, got, tc.want)
			}
		})
	}
}

func TestRequireAnyPermission(t *testing.T) {
	registry := newTestRegistry(t)

	tests := []struct {
		name         string
		injectClaims *Claims
		required     []models.Permission
		wantStatus   int
		wantErrorMsg string
	}{
		{
			name:         "allows request when user holds required permission",
			injectClaims: &Claims{UserID: 1, Roles: []models.RoleName{models.RoleCustomer}},
			required:     []models.Permission{models.UserReadOwn, models.AccountReadOwn},
			wantStatus:   http.StatusOK,
		},
		{
			name:         "forbids request when user lacks required permission",
			injectClaims: &Claims{UserID: 2, Roles: []models.RoleName{models.RoleCustomer}},
			required:     []models.Permission{models.UserReadAll},
			wantStatus:   http.StatusForbidden,
			wantErrorMsg: "forbidden: insufficient permissions",
		},
		{
			name:         "returns 401 when claims are missing from context",
			injectClaims: nil,
			required:     []models.Permission{models.UserReadOwn},
			wantStatus:   http.StatusUnauthorized,
			wantErrorMsg: "unauthorized",
		},
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/users/profile", nil)
			if tc.injectClaims != nil {
				req = req.WithContext(WithClaims(req.Context(), tc.injectClaims))
			}

			rec := httptest.NewRecorder()
			handler := RequireAnyPermission(registry, tc.required...)(nextHandler)
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("got status %d, want %d", rec.Code, tc.wantStatus)
			}

			if tc.wantStatus != http.StatusOK {
				var errResp map[string]string
				if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
					t.Fatalf("failed to unmarshal error JSON: %v", err)
				}
				if errResp["error"] != tc.wantErrorMsg {
					t.Errorf("got error %q, want %q", errResp["error"], tc.wantErrorMsg)
				}
			}
		})
	}
}
