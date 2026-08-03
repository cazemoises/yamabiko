package srs

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

func (r *PostgresRepository) Get(ctx context.Context, userID, chunkID uuid.UUID) (Card, Status, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT easiness_factor, interval_days, repetitions, status
		 FROM user_chunk_progress WHERE user_id = $1 AND chunk_id = $2`,
		userID, chunkID,
	)
	var card Card
	var status string
	err := row.Scan(&card.EasinessFactor, &card.IntervalDays, &card.Repetitions, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Card{}, StatusNovo, nil
	}
	if err != nil {
		return Card{}, "", err
	}
	return card, Status(status), nil
}

func (r *PostgresRepository) Save(ctx context.Context, userID, chunkID uuid.UUID, review Review) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_chunk_progress (user_id, chunk_id, status, easiness_factor, interval_days, repetitions, next_review_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (user_id, chunk_id) DO UPDATE SET
		   status = $3, easiness_factor = $4, interval_days = $5, repetitions = $6, next_review_at = $7`,
		userID, chunkID, string(review.Status), review.Card.EasinessFactor, review.Card.IntervalDays, review.Card.Repetitions, review.NextReviewAt,
	)
	return err
}

func (r *PostgresRepository) CountByStatus(ctx context.Context, userID uuid.UUID) (map[Status]int, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT status, count(*) FROM user_chunk_progress WHERE user_id = $1 GROUP BY status`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[Status]int{}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		result[Status(status)] = count
	}
	return result, rows.Err()
}
