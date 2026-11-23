package postgres

import (
	"context"

	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionRepository struct {
	db *pgxpool.Pool
}

func NewPermissionRepository(db *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) LoadRolePermissions(ctx context.Context) (map[models.RoleName]map[models.Permission]struct{}, error) {
	query := `
		SELECT r.name, p.name
		FROM role_permissions rp
		JOIN roles r ON r.id = rp.role_id
		JOIN permissions p ON p.id = rp.permissions_id
    `
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make(
		map[models.RoleName]map[models.Permission]struct{},
	)

	for rows.Next() {
		var role models.RoleName
		var permission models.Permission

		if err = rows.Scan(&role, &permission); err != nil {
			return nil, err
		}

		if _, exists := permissions[role]; !exists {
			permissions[role] = make(
				map[models.Permission]struct{},
			)
		}

		permissions[role][permission] = struct{}{}
	}

	return permissions, rows.Err()
}
