package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/fouched/novabank/users/internal/core/models"
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
	// Idiomatic Go SQL: Pull User info and aggregate associated Roles array in a single query
	query := `
		SELECT u.id, u.email, u.password, u.first_name, u.last_name, u.enabled, u.created_at,
		       COALESCE(json_agg(json_build_object('id', r.id, 'name', r.name)) FILTER (WHERE r.id IS NOT NULL), '[]')
		FROM users u
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id
		WHERE u.email = $1
		GROUP BY u.id`

	var user models.User
	var rolesJSON []models.Role

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Enabled, &user.CreatedAt, &rolesJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to extract user by email: %w", err)
	}
	user.Roles = rolesJSON
	return &user, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, email).Scan(&exists); err != nil {
		return false, fmt.Errorf("error verifying user existence: %w", err)
	}
	return exists, nil
}

func (r *UserRepository) CountByEnabledTrue(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(1) FROM users WHERE enabled = TRUE`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("error counting active users: %w", err)
	}
	return count, nil
}

func (r *UserRepository) FindByEnabledTrue(ctx context.Context) ([]models.User, error) {
	query := `SELECT id, email, password, first_name, last_name, enabled, created_at FROM users WHERE enabled = TRUE`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error fetching active users list: %w", err)
	}
	defer rows.Close()

	var activeUsers []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.Enabled, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning user row data: %w", err)
		}
		activeUsers = append(activeUsers, u)
	}
	return activeUsers, nil
}

func (r *UserRepository) CountByRoleName(ctx context.Context, roleName string) (int64, error) {
	query := `SELECT COUNT(DISTINCT u.id) FROM users u 
	          JOIN user_roles ur ON u.id = ur.user_id 
	          JOIN roles r ON ur.role_id = r.id 
	          WHERE r.name = $1`
	var count int64
	if err := r.db.QueryRow(ctx, query, roleName).Scan(&count); err != nil {
		return 0, fmt.Errorf("error counting users by role criteria: %w", err)
	}
	return count, nil
}

func (r *UserRepository) FindByRoleName(ctx context.Context, roleName string) ([]models.User, error) {
	query := `SELECT DISTINCT u.id, u.email, u.password, u.first_name, u.last_name, u.enabled, u.created_at 
	          FROM users u 
	          JOIN user_roles ur ON u.id = ur.user_id 
	          JOIN roles r ON ur.role_id = r.id 
	          WHERE r.name = $1`
	rows, err := r.db.Query(ctx, query, roleName)
	if err != nil {
		return nil, fmt.Errorf("error searching users filtered by role allocation: %w", err)
	}
	defer rows.Close()

	var matchingUsers []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.Enabled, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("error parsing user row query context: %w", err)
		}
		matchingUsers = append(matchingUsers, u)
	}
	return matchingUsers, nil
}

func (r *UserRepository) Save(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (email, password, first_name, last_name, enabled) 
	          VALUES ($1, $2, $3, $4, $5) 
	          RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, user.Email, user.Password, user.FirstName, user.LastName, user.Enabled).Scan(&user.ID, &user.CreatedAt)
}
