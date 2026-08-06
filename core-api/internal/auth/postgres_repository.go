package auth

import (
	"context"
	"errors"
	"time"

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
	return r.scanUser(ctx, `SELECT id, email, password_hash, name, created_at, current_sprint_day,
	                                accent_color, pin_hash, pin_failed_attempts, pin_locked_until
	                         FROM users WHERE email = $1`, email)
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return r.scanUser(ctx, `SELECT id, email, password_hash, name, created_at, current_sprint_day,
	                                accent_color, pin_hash, pin_failed_attempts, pin_locked_until
	                         FROM users WHERE id = $1`, id)
}

func (r *PostgresRepository) scanUser(ctx context.Context, query string, arg any) (*User, error) {
	row := r.pool.QueryRow(ctx, query, arg)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt, &u.CurrentSprintDay,
		&u.AccentColor, &u.PinHash, &u.PinFailedAttempts, &u.PinLockedUntil)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ListPinEnabledProfiles alimenta GET /auth/profiles — projeção mínima
// (id, name, accent_color), nunca email/pin_hash. Ordenado por nome pra
// UI ter ordem estável entre requests.
func (r *PostgresRepository) ListPinEnabledProfiles(ctx context.Context) ([]PinProfile, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, accent_color FROM users WHERE pin_hash IS NOT NULL ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]PinProfile, 0)
	for rows.Next() {
		var p PinProfile
		if err := rows.Scan(&p.ID, &p.DisplayName, &p.AccentColor); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

func (r *PostgresRepository) UpdatePinAuthState(ctx context.Context, userID uuid.UUID, failedAttempts int, lockedUntil *time.Time) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET pin_failed_attempts = $2, pin_locked_until = $3 WHERE id = $1`,
		userID, failedAttempts, lockedUntil,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *PostgresRepository) SetPinHash(ctx context.Context, userID uuid.UUID, hash string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE users SET pin_hash = $2 WHERE id = $1`, userID, hash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
