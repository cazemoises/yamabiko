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

// FindOrCreateByEmail resolve o usuário pelo email do header Remote-Email
// do Pangolin — cria na hora se for o primeiro acesso dessa pessoa (SSO por
// pessoa, sem fluxo de registro separado; ver context.go). `name` só é
// usado na criação — mudanças de nome no Pangolin depois não retroagem pro
// yamabiko.
func (r *PostgresRepository) FindOrCreateByEmail(ctx context.Context, email, name string) (uuid.UUID, error) {
	row := r.pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email)
	var id uuid.UUID
	err := row.Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}

	id = uuid.New()
	_, err = r.pool.Exec(ctx,
		`INSERT INTO users (id, email, name, current_sprint_day) VALUES ($1, $2, $3, 1)`,
		id, email, name,
	)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}
