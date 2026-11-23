package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/fouched/novabank/users/internal/core/errdefs"
	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/fouched/toolkit/v2/faults"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error) {
	query := `
		SELECT a.id, a.account_number, a.balance, a.currency, a.account_type, a.account_status, a.user_id, a.created_at, u.email 
		FROM accounts a
		JOIN users u ON u.id = a.user_id
		WHERE a.account_number = $1`

	var acc models.Account
	err := r.db.QueryRow(ctx, query, accountNumber).Scan(
		&acc.ID, &acc.AccountNumber, &acc.Balance, &acc.Currency,
		&acc.AccountType, &acc.AccountStatus, &acc.UserID, &acc.CreatedAt, &acc.UserEmail,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errdefs.NewNotFound("account not found")
		}
		return nil, faults.WithStack(fmt.Errorf("error scanning account by number: %w", err))
	}
	return &acc, nil
}

func (r *AccountRepository) ExistsByAccountNumber(ctx context.Context, accountNumber string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM accounts WHERE account_number = $1)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, accountNumber).Scan(&exists); err != nil {
		return false, faults.WithStack(fmt.Errorf("error verifying account existence: %w", err))
	}
	return exists, nil
}

func (r *AccountRepository) GetByUserID(ctx context.Context, userID int64) (*models.Account, error) {
	query := `
		SELECT a.id, a.account_number, a.balance, a.currency, a.account_type, a.account_status, a.user_id, a.created_at, u.email 
		FROM accounts a
		JOIN users u ON u.id = a.user_id
		WHERE u.id = $1
		ORDER BY a.created_at DESC`

	var acc models.Account
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&acc.ID, &acc.AccountNumber, &acc.Balance, &acc.Currency,
		&acc.AccountType, &acc.AccountStatus, &acc.UserID, &acc.CreatedAt, &acc.UserEmail,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errdefs.ErrNotFound
		}
		return nil, faults.WithStack(fmt.Errorf("error scanning account by user id: %w", err))
	}
	return &acc, nil
}

func (r *AccountRepository) GetByEmail(ctx context.Context, email string) ([]models.Account, error) {
	query := `
		SELECT a.id, a.account_number, a.balance, a.currency, a.account_type, a.account_status, a.user_id, a.created_at, u.email
		FROM accounts a
		JOIN users u ON u.id = a.user_id
		WHERE u.email = $1
		ORDER BY a.created_at DESC`

	rows, err := r.db.Query(ctx, query, email)
	if err != nil {
		return nil, err
	}

	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		err = rows.Scan(&a.ID, &a.AccountNumber, &a.Balance, &a.Currency,
			&a.AccountType, &a.AccountStatus, &a.UserID, &a.CreatedAt, &a.UserEmail,
		)
		if err != nil {
			return nil, faults.WithStack(err)
		}
		accounts = append(accounts, a)
	}

	if err = rows.Err(); err != nil {
		return nil, faults.WithStack(err)
	}

	if len(accounts) == 0 {
		return nil, errdefs.ErrNotFound
	}

	return accounts, nil
}

func (r *AccountRepository) Insert(ctx context.Context, account *models.Account) error {
	query := `INSERT INTO accounts (account_number, balance, currency, account_type, account_status, user_id) 
	          VALUES ($1, $2, $3, $4, $5, $6) 
	          RETURNING id, created_at`
	err := r.db.QueryRow(ctx, query,
		account.AccountNumber, account.Balance, account.Currency,
		account.AccountType, account.AccountStatus, account.UserID,
	).Scan(&account.ID, &account.CreatedAt)

	return faults.WithStack(err)
}

func (r *AccountRepository) CountAll(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM accounts`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, faults.WithStack(err)
	}
	return count, nil
}

func (r *AccountRepository) ChangeAccountStatus(ctx context.Context, accountNumber string, status models.AccountStatus) (bool, error) {
	query := `
		UPDATE accounts 
		SET account_status = $1 
		WHERE account_number = $2`

	cmdTag, err := r.db.Exec(ctx, query, status, accountNumber)
	if err != nil {
		return false, faults.WithStack(err)
	}

	if cmdTag.RowsAffected() == 0 {
		return false, faults.WithStack(errdefs.NewNotFound("account not found"))
	}

	return true, nil
}

func (r *AccountRepository) GetAll(ctx context.Context, pageSize, page int) ([]models.Account, error) {
	query := `
		SELECT a.id, a.account_number, a.balance, a.currency, a.account_type, a.account_status, a.user_id, a.created_at, u.email
		FROM accounts a
		JOIN users u ON u.id = a.user_id
		ORDER BY a.created_at DESC
		LIMIT $1 
		OFFSET $2`

	rows, err := r.db.Query(ctx, query, pageSize, page)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		err = rows.Scan(&a.ID, &a.AccountNumber, &a.Balance, &a.Currency,
			&a.AccountType, &a.AccountStatus, &a.UserID, &a.CreatedAt, &a.UserEmail,
		)
		if err != nil {
			return nil, faults.WithStack(err)
		}
		accounts = append(accounts, a)
	}

	if err = rows.Err(); err != nil {
		return nil, faults.WithStack(err)
	}

	if len(accounts) == 0 {
		return nil, errdefs.ErrNotFound
	}

	return accounts, nil
}
