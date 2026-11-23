package repositories

import (
	"context"

	"github.com/fouched/novabank/users/internal/core/models"
)

type PermissionRepository interface {
	LoadRolePermissions(ctx context.Context) (map[models.RoleName]map[models.Permission]struct{}, error)
}
