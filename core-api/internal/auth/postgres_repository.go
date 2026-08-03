package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, user *User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, name, current_sprint_day) VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Email, user.PasswordHash, user.Name, user.CurrentSprintDay,
	)
	return err
}

func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	return r.scanUser(ctx, `SELECT id, email, password_hash, name, created_at, current_sprint_day FROM users WHERE email = $1`, email)
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return r.scanUser(ctx, `SELECT id, email, password_hash, name, created_at, current_sprint_day FROM users WHERE id = $1`, id)
}

func (r *PostgresRepository) scanUser(ctx context.Context, query string, arg any) (*User, error) {
	row := r.pool.QueryRow(ctx, query, arg)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt, &u.CurrentSprintDay)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
