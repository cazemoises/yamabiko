package scenarios

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

func (r *PostgresRepository) List(ctx context.Context, filter Filter) ([]Scenario, error) {
	query := `SELECT id, language, title_pt, context_description_pt, order_index FROM scenarios`

	var conditions []string
	var args []any
	if filter.Language != nil {
		args = append(args, *filter.Language)
		conditions = append(conditions, fmt.Sprintf("language = $%d", len(args)))
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY order_index"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []Scenario{}
	for rows.Next() {
		s, err := scanScenario(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*Scenario, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, language, title_pt, context_description_pt, order_index FROM scenarios WHERE id = $1`,
		id,
	)
	s, err := scanScenario(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrScenarioNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanScenario(row rowScanner) (Scenario, error) {
	var s Scenario
	err := row.Scan(&s.ID, &s.Language, &s.TitlePT, &s.ContextDescriptionPT, &s.OrderIndex)
	if err != nil {
		return Scenario{}, err
	}
	return s, nil
}
