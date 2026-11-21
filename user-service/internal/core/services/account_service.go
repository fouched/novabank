package services

import (
	"context"
	"log/slog"

	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/fouched/novabank/users/internal/repositories"
)

type AccountService struct {
	accountRepo repositories.AccountRepository
	logger      *slog.Logger
}

func NewAccountService(
	repo repositories.AccountRepository,
	logger *slog.Logger,
) *AccountService {
	return &AccountService{
		accountRepo: repo,
		logger:      logger,
	}
}

func (s *AccountService) GetByUserID(ctx context.Context, userID int64) (*models.Account, error) {
	account, err := s.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *AccountService) GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error) {
	account, err := s.accountRepo.GetByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, err
	}
	return account, nil
}

// GetByAccountNumberAndEmail gets and account by number and email
// ensuring the user actually owns the account
func (s *AccountService) GetByAccountNumberAndEmail(ctx context.Context, accountNumber, email string) (*models.Account, error) {
	account, err := s.accountRepo.GetByAccountNumberAndEmail(ctx, accountNumber, email)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *AccountService) GetByEmail(ctx context.Context, email string) ([]models.Account, error) {
	accounts, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *AccountService) ChangeAccountStatus(ctx context.Context, accountNumber string, status models.AccountStatus) (bool, error) {
	hasStatusChanged, err := s.accountRepo.ChangeAccountStatus(ctx, accountNumber, status)
	if err != nil {
		return false, err
	}
	return hasStatusChanged, nil
}

func (s *AccountService) GetAll(ctx context.Context, pageSize, page int) ([]models.Account, error) {
	accounts, err := s.accountRepo.GetAll(ctx, pageSize, page)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}
