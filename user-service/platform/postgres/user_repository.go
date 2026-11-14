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

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

const userSelectBase = `
	SELECT
		u.id,
		u.email,
		u.password,
		u.first_name,
		u.last_name,
		u.enabled,
		u.created_at,
		(
			SELECT COALESCE(
				json_agg(
					json_build_object(
						'id', r.id,
						'name', r.name
					)
				),
				'[]'::json
			)
			FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = u.id
		) AS roles,
		(
			SELECT COALESCE(
				json_agg(
					json_build_object(
						'id', a.id,
						'account_number', a.account_number,
						'balance', a.balance,
						'currency', a.currency,
						'account_type', a.account_type,
						'account_status', a.account_status,
						'created_at', a.created_at
					)
				),
				'[]'::json
			)
			FROM accounts a
			WHERE a.user_id = u.id
		) AS accounts
	FROM users u
`

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := r.findOne(ctx, "WHERE u.email = $1", email)
	if err != nil {
		if errors.Is(err, errdefs.ErrNotFound) {
			return nil, err
		}
		return nil, faults.WithStack(fmt.Errorf("find user by email: %w", err))
	}

	return user, nil
}

func (r *UserRepository) FindByAccountNumber(ctx context.Context, accountNumber string) (*models.User, error) {
	const whereClause = `
		WHERE EXISTS (
			SELECT 1
			FROM accounts a
			WHERE a.user_id = u.id
			  AND a.account_number = $1
		)
	`

	user, err := r.findOne(ctx, whereClause, accountNumber)
	if err != nil {
		if errors.Is(err, errdefs.ErrNotFound) {
			return nil, err
		}
		return nil, faults.WithStack(fmt.Errorf("find user by account number: %w", err))
	}

	return user, nil
}

func (r *UserRepository) findOne(ctx context.Context, whereClause string, args ...any) (*models.User, error) {
	query := userSelectBase + "\n" + whereClause

	var user models.User

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Enabled,
		&user.CreatedAt,
		&user.Roles,
		&user.Accounts,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errdefs.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, email).Scan(&exists); err != nil {
		return false, faults.WithStack(fmt.Errorf("error verifying user existence: %w", err))
	}
	return exists, nil
}

// Insert runs in a transaction and creates the user, user role and account
func (r *UserRepository) Insert(ctx context.Context, user *models.User) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO users (email, password, first_name, last_name, enabled) 
	          VALUES ($1, $2, $3, $4, $5) 
	          RETURNING id, created_at`
	err = tx.QueryRow(ctx, query, user.Email, user.Password, user.FirstName, user.LastName, user.Enabled).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return faults.WithStack(err)
	}

	query = `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`
	_, err = tx.Exec(ctx, query, user.ID, user.Roles[0].ID)

	if err != nil {
		return faults.WithStack(err)
	}

	query = `INSERT INTO accounts (account_number, balance, currency, account_type, account_status, user_id) 
	          VALUES ($1, $2, $3, $4, $5, $6) 
	          RETURNING id, created_at`
	err = tx.QueryRow(ctx, query,
		user.Accounts[0].AccountNumber, user.Accounts[0].Balance, user.Accounts[0].Currency,
		user.Accounts[0].AccountType, user.Accounts[0].AccountStatus, user.ID,
	).Scan(&user.Accounts[0].ID, &user.Accounts[0].CreatedAt)

	if err != nil {
		return faults.WithStack(err)
	}

	user.Accounts[0].UserID = user.ID

	return tx.Commit(ctx)
}

func (r *UserRepository) CountAll(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, faults.WithStack(err)
	}
	return count, nil
}

func (r *UserRepository) CountByEnabled(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users
			WHERE enabled IS TRUE`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, faults.WithStack(err)
	}
	return count, nil
}

func (r *UserRepository) CountByRoleName(ctx context.Context, roleName models.RoleName) (int64, error) {
	query := `SELECT COUNT(u.*) FROM users u
            JOIN user_roles ur ON u.id = ur.user_id
            JOIN roles r ON r.id = ur.role_id      
			WHERE r.name = $1`
	var count int64
	if err := r.db.QueryRow(ctx, query, roleName).Scan(&count); err != nil {
		return 0, faults.WithStack(err)
	}
	return count, nil
}

func (r *UserRepository) FindByRoleName(ctx context.Context, roleName string, pageSize, page int) ([]models.User, error) {
	query := `SELECT
				u.id,
				u.email,
				u.first_name,
				u.last_name,
				u.enabled,
				u.created_at    
			FROM users u
			JOIN user_roles ur ON ur.user_id = u.id
			JOIN roles r ON r.id = ur.role_id
			WHERE r.name = $1
			ORDER BY u.created_at DESC
			LIMIT $2 
			OFFSET $3`

	rows, err := r.db.Query(ctx, query, roleName, pageSize, page)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err = rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.Enabled, &u.CreatedAt)
		if err != nil {
			return nil, faults.WithStack(err)
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, faults.WithStack(err)
	}

	if len(users) == 0 {
		return nil, errdefs.ErrNotFound
	}

	return users, nil
}

func (r *UserRepository) ToggleUserStatus(ctx context.Context, userID int64) (bool, error) {
	query := "UPDATE users SET enabled = NOT enabled WHERE id = $1 RETURNING enabled"

	var enabled bool
	if err := r.db.QueryRow(ctx, query, userID).Scan(&enabled); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, errdefs.ErrNotFound
		}
	}

	return enabled, nil
}
