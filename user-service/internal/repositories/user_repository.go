package repositories

import (
	"context"

	"github.com/fouched/novabank/users/internal/core/models"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Insert(ctx context.Context, user *models.User) error
}
