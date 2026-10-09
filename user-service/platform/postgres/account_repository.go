package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) FindByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error) {
	query := `SELECT id, account_number, balance, currency, account_type, account_status, user_id, created_at 
	          FROM accounts WHERE account_number = $1`

	var acc models.Account
	err := r.db.QueryRow(ctx, query, accountNumber).Scan(
		&acc.ID, &acc.AccountNumber, &acc.Balance, &acc.Currency,
		&acc.AccountType, &acc.AccountStatus, &acc.UserID, &acc.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error scanning account by number: %w", err)
	}
	return &acc, nil
}

func (r *AccountRepository) ExistsByAccountNumber(ctx context.Context, accountNumber string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM accounts WHERE account_number = $1)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, accountNumber).Scan(&exists); err != nil {
		return false, fmt.Errorf("error verifying account existence: %w", err)
	}
	return exists, nil
}

func (r *AccountRepository) FindByUserID(ctx context.Context, userID int64) (*models.Account, error) {
	query := `SELECT id, account_number, balance, currency, account_type, account_status, user_id, created_at 
	          FROM accounts WHERE user_id = $1`

	var acc models.Account
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&acc.ID, &acc.AccountNumber, &acc.Balance, &acc.Currency,
		&acc.AccountType, &acc.AccountStatus, &acc.UserID, &acc.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error scanning account by user id: %w", err)
	}
	return &acc, nil
}

func (r *AccountRepository) Save(ctx context.Context, account *models.Account) error {
	query := `INSERT INTO accounts (account_number, balance, currency, account_type, account_status, user_id) 
	          VALUES ($1, $2, $3, $4, $5, $6) 
	          RETURNING id, created_at`
	return r.db.QueryRow(ctx, query,
		account.AccountNumber, account.Balance, account.Currency,
		account.AccountType, account.AccountStatus, account.UserID,
	).Scan(&account.ID, &account.CreatedAt)
}
