package phonetics

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yamabiko/core-api/internal/comparison"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) IncrementPatterns(ctx context.Context, userID uuid.UUID, language string, counts map[comparison.ErrorPattern]int) error {
	for pattern, count := range counts {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO phonetic_error_patterns (user_id, pattern_type, language, occurrences, last_seen_at)
			 VALUES ($1, $2, $3, $4, now())
			 ON CONFLICT (user_id, pattern_type, language)
			 DO UPDATE SET occurrences = phonetic_error_patterns.occurrences + $4, last_seen_at = now()`,
			userID, string(pattern), language, count,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) Heatmap(ctx context.Context, userID uuid.UUID) ([]PatternCount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT pattern_type, occurrences FROM phonetic_error_patterns WHERE user_id = $1 ORDER BY occurrences DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []PatternCount{}
	for rows.Next() {
		var pc PatternCount
		var patternType string
		if err := rows.Scan(&patternType, &pc.Occurrences); err != nil {
			return nil, err
		}
		pc.Pattern = comparison.ErrorPattern(patternType)
		result = append(result, pc)
	}
	return result, rows.Err()
}
