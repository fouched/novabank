package repositories

import (
	"context"

	"github.com/fouched/novabank/users/internal/core/models"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Insert(ctx context.Context, user *models.User) error
	CountAll(ctx context.Context) (int64, error)
	CountByEnabled(ctx context.Context) (int64, error)
	CountByRoleName(ctx context.Context, roleName models.RoleName) (int64, error)
	GetByRoleName(ctx context.Context, roleName string, pageSize, page int) ([]models.User, error)
	ToggleUserStatus(ctx context.Context, userID int64) (bool, error)
}
