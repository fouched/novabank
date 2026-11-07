package postgres

import (
	"context"
	"encoding/json"
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

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `
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
			) AS roles
		FROM users u
		WHERE u.email = $1
	`

	var user models.User
	var rolesJSON []byte

	row := r.db.QueryRow(ctx, query, email)

	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Enabled, &user.CreatedAt, &rolesJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, faults.WithStack(errdefs.NewNotFound("user not found"))
		}

		return nil, faults.WithStack(fmt.Errorf("find user by email: %w", err))
	}

	if err := json.Unmarshal(rolesJSON, &user.Roles); err != nil {
		return nil, faults.WithStack(fmt.Errorf("unmarshal user roles: %w", err))
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
	err = tx.QueryRow(ctx, query, user.Email, user.Password, user.FirstName, user.LastName, user.Enabled).Scan(&user.ID, &user.CreatedAt)
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
