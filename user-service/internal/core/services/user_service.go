package services

import (
	"context"
	"log/slog"
	"strings"

	"github.com/fouched/novabank/users/internal/core/errdefs"
	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/fouched/novabank/users/internal/repositories"
)

type UserService struct {
	userRepo    repositories.UserRepository
	accountRepo repositories.AccountRepository
	logger      *slog.Logger
}

func NewUserService(
	uRepo repositories.UserRepository,
	aRepo repositories.AccountRepository,
	logger *slog.Logger,
) *UserService {
	return &UserService{
		userRepo:    uRepo,
		accountRepo: aRepo,
		logger:      logger,
	}
}

func (s *UserService) GetUserAccountDetails(ctx context.Context, email string) (*models.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Search(ctx context.Context, email, accountNumber string) (*models.User, error) {
	email = strings.TrimSpace(email)
	accountNumber = strings.TrimSpace(accountNumber)
	if email != "" {
		u, err := s.userRepo.FindByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		return u, nil
	} else if strings.TrimSpace(accountNumber) != "" {
		u, err := s.userRepo.FindByAccountNumber(ctx, accountNumber)
		if err != nil {
			return nil, err
		}
		return u, nil
	}

	return nil, errdefs.NewNotFound("resource not found")
}

func (s *UserService) GetAllUsersByRole(ctx context.Context, roleName string, pageSize, page int) ([]models.User, error) {
	role := strings.TrimSpace(roleName)
	users, err := s.userRepo.FindByRoleName(ctx, role, pageSize, page)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UserService) GetUserStats(ctx context.Context) (models.UserStats, error) {
	total, err := s.userRepo.CountAll(ctx)
	if err != nil {
		return models.UserStats{}, err
	}
	enabled, err := s.userRepo.CountByEnabled(ctx)
	if err != nil {
		return models.UserStats{}, err
	}
	totalAccounts, err := s.accountRepo.CountAll(ctx)
	if err != nil {
		return models.UserStats{}, err
	}
	customerCount, err := s.userRepo.CountByRoleName(ctx, models.RoleCustomer)
	if err != nil {
		return models.UserStats{}, err
	}
	adminCount, err := s.userRepo.CountByRoleName(ctx, models.RoleAdmin)
	if err != nil {
		return models.UserStats{}, err
	}

	return models.UserStats{
		TotalUsers:    total,
		ActiveUsers:   enabled,
		InactiveUsers: total - enabled,
		TotalAccounts: totalAccounts,
		CustomerCount: customerCount,
		AdminCount:    adminCount,
	}, nil
}

func (s *UserService) ToggleUserStatus(ctx context.Context, userID int64) (bool, error) {
	return s.userRepo.ToggleUserStatus(ctx, userID)
}
