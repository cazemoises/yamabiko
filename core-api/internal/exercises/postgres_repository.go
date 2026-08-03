package exercises

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

func (r *PostgresRepository) List(ctx context.Context, filter Filter) ([]Exercise, error) {
	query := `SELECT id, category, difficulty, prompt_pt, expected_transcript, expected_romaji, sprint_day_ref FROM exercises`

	var conditions []string
	var args []any

	if filter.SprintDay != nil {
		args = append(args, *filter.SprintDay)
		conditions = append(conditions, fmt.Sprintf("sprint_day_ref = $%d", len(args)))
	}
	if filter.Category != nil {
		args = append(args, *filter.Category)
		conditions = append(conditions, fmt.Sprintf("category = $%d", len(args)))
	}
	if filter.Difficulty != nil {
		args = append(args, *filter.Difficulty)
		conditions = append(conditions, fmt.Sprintf("difficulty = $%d", len(args)))
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY sprint_day_ref, category"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []Exercise{}
	for rows.Next() {
		e, err := scanExercise(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*Exercise, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, category, difficulty, prompt_pt, expected_transcript, expected_romaji, sprint_day_ref
		 FROM exercises WHERE id = $1`,
		id,
	)
	e, err := scanExercise(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrExerciseNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanExercise(row rowScanner) (Exercise, error) {
	var e Exercise
	var romaji *string
	err := row.Scan(&e.ID, &e.Category, &e.Difficulty, &e.PromptPT, &e.ExpectedTranscript, &romaji, &e.SprintDayRef)
	if err != nil {
		return Exercise{}, err
	}
	if romaji != nil {
		e.ExpectedRomaji = *romaji
	}
	return e, nil
}
