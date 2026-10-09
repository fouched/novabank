package repositories

import (
	"context"

	"github.com/fouched/novabank/users/internal/core/models"
)

type AccountRepository interface {
	FindByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error)
	ExistsByAccountNumber(ctx context.Context, accountNumber string) (bool, error)
	FindByUserID(ctx context.Context, userID int64) (*models.Account, error)
	Save(ctx context.Context, account *models.Account) error
}
