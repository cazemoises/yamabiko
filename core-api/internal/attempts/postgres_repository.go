package attempts

import (
	"context"
	"encoding/json"

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

func (r *PostgresRepository) Create(ctx context.Context, attempt *Attempt) error {
	diffJSON, err := json.Marshal(attempt.PhoneticDiff)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO attempts (id, user_id, exercise_id, audio_transcript, similarity_score, verdict, phonetic_diff)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		attempt.ID, attempt.UserID, attempt.ExerciseID, attempt.AudioTranscript,
		attempt.SimilarityScore, string(attempt.Verdict), json.RawMessage(diffJSON),
	)
	return err
}

func (r *PostgresRepository) ListByUserAndExercise(ctx context.Context, userID, exerciseID uuid.UUID) ([]Attempt, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, exercise_id, audio_transcript, similarity_score, verdict, phonetic_diff, created_at
		 FROM attempts WHERE user_id = $1 AND exercise_id = $2 ORDER BY created_at DESC`,
		userID, exerciseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []Attempt{}
	for rows.Next() {
		var a Attempt
		var diffJSON []byte
		var verdict string
		if err := rows.Scan(&a.ID, &a.UserID, &a.ExerciseID, &a.AudioTranscript, &a.SimilarityScore, &verdict, &diffJSON, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Verdict = comparison.Verdict(verdict)
		if err := json.Unmarshal(diffJSON, &a.PhoneticDiff); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}
