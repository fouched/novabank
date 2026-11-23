package repositories

import (
	"context"

	"github.com/fouched/novabank/users/internal/core/models"
)

type AccountRepository interface {
	GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error)
	ExistsByAccountNumber(ctx context.Context, accountNumber string) (bool, error)
	GetByUserID(ctx context.Context, userID int64) (*models.Account, error)
	GetByEmail(ctx context.Context, email string) ([]models.Account, error)
	Insert(ctx context.Context, account *models.Account) error
	CountAll(ctx context.Context) (int64, error)
	ChangeAccountStatus(ctx context.Context, accountNumber string, status models.AccountStatus) (bool, error)
	GetAll(ctx context.Context, pageSize, page int) ([]models.Account, error)
}
