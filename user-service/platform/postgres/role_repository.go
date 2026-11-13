package postgres

import (
	"context"
	"errors"

	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/fouched/toolkit/v2/faults"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindByName(ctx context.Context, name models.RoleName) (*models.Role, error) {
	query := `SELECT id, name FROM roles WHERE name = $1`

	var role models.Role
	err := r.db.QueryRow(ctx, query, name).Scan(&role.ID, &role.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil, nil when entity isn't found (Go standard practice for Optionals)
		}
		return nil, faults.WithStack(err)
	}
	return &role, nil
}
