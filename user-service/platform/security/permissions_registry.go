package security

import (
	"context"
	"sync"

	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/fouched/novabank/users/internal/repositories"
)

type PermissionRegistry struct {
	repo        repositories.PermissionRepository
	permissions map[models.RoleName]map[models.Permission]struct{}
	mu          sync.RWMutex
}

func NewPermissionRegistry(repo repositories.PermissionRepository) *PermissionRegistry {
	return &PermissionRegistry{
		repo:        repo,
		permissions: make(map[models.RoleName]map[models.Permission]struct{}),
	}
}

func (r *PermissionRegistry) Reload(ctx context.Context) error {
	permissions, err := r.repo.LoadRolePermissions(ctx)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.permissions = permissions

	return nil
}

func (r *PermissionRegistry) HasAnyPermission(roles []models.RoleName, permissions ...models.Permission) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, role := range roles {

		rolePermissions, ok := r.permissions[role]
		if !ok {
			continue
		}

		for _, permission := range permissions {

			if _, exists := rolePermissions[permission]; exists {
				return true
			}
		}
	}

	return false
}
