package repositories

import (
	"context"

	"github.com/fouched/novabank/users/internal/core/models"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	CountByEnabledTrue(ctx context.Context) (int64, error)
	FindByEnabledTrue(ctx context.Context) ([]models.User, error)
	CountByRoleName(ctx context.Context, roleName string) (int64, error)
	FindByRoleName(ctx context.Context, roleName string) ([]models.User, error)
	Save(ctx context.Context, user *models.User) error
}
