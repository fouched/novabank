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
	account, err := s.accountRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *AccountService) GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error) {
	account, err := s.accountRepo.FindByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *AccountService) GetByEmail(ctx context.Context, email string) (*models.Account, error) {
	account, err := s.accountRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *AccountService) ChangeAccountStatus(ctx context.Context, accountNumber string, status models.AccountStatus) (bool, error) {
	return false, nil
}
