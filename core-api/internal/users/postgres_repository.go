package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yamabiko/core-api/internal/gamification"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) FindProfileByID(ctx context.Context, id uuid.UUID) (*Profile, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, name, created_at, current_sprint_day, xp_total, current_streak_days, longest_streak_days, last_attempt_date
		 FROM users WHERE id = $1`,
		id,
	)
	var p Profile
	err := row.Scan(&p.ID, &p.Email, &p.Name, &p.CreatedAt, &p.CurrentSprintDay,
		&p.XPTotal, &p.CurrentStreakDays, &p.LongestStreakDays, &p.LastAttemptDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	badges, err := r.EarnedBadges(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Badges = make([]string, 0, len(badges))
	for badge := range badges {
		p.Badges = append(p.Badges, string(badge))
	}

	return &p, nil
}

func (r *PostgresRepository) FindStatsByID(ctx context.Context, id uuid.UUID) (*gamification.UserStats, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT xp_total, current_streak_days, longest_streak_days, last_attempt_date FROM users WHERE id = $1`,
		id,
	)
	var stats gamification.UserStats
	err := row.Scan(&stats.XPTotal, &stats.CurrentStreakDays, &stats.LongestStreakDays, &stats.LastAttemptDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *PostgresRepository) UpdateStats(ctx context.Context, userID uuid.UUID, stats gamification.UserStats) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET xp_total = $2, current_streak_days = $3, longest_streak_days = $4, last_attempt_date = $5
		 WHERE id = $1`,
		userID, stats.XPTotal, stats.CurrentStreakDays, stats.LongestStreakDays, stats.LastAttemptDate,
	)
	return err
}

func (r *PostgresRepository) EarnedBadges(ctx context.Context, userID uuid.UUID) (map[gamification.Badge]bool, error) {
	rows, err := r.pool.Query(ctx, `SELECT badge_code FROM user_badges WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[gamification.Badge]bool{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		result[gamification.Badge(code)] = true
	}
	return result, rows.Err()
}

func (r *PostgresRepository) AwardBadges(ctx context.Context, userID uuid.UUID, badges []gamification.Badge) error {
	for _, badge := range badges {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO user_badges (user_id, badge_code) VALUES ($1, $2) ON CONFLICT (user_id, badge_code) DO NOTHING`,
			userID, string(badge),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
