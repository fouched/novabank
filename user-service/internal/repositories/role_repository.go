package repositories

import (
	"context"

	"github.com/fouched/novabank/users/internal/core/models"
)

type RoleRepository interface {
	FindByName(ctx context.Context, name string) (*models.Role, error)
}
