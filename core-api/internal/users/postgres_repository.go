package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		`SELECT id, email, name, created_at, current_sprint_day, xp_total, current_streak_days, longest_streak_days,
		        last_attempt_date, preferred_voice_ja, preferred_voice_en, theme, accent_color
		 FROM users WHERE id = $1`,
		id,
	)
	var p Profile
	err := row.Scan(&p.ID, &p.Email, &p.Name, &p.CreatedAt, &p.CurrentSprintDay,
		&p.XPTotal, &p.CurrentStreakDays, &p.LongestStreakDays, &p.LastAttemptDate,
		&p.PreferredVoiceJA, &p.PreferredVoiceEN, &p.Theme, &p.AccentColor)
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

// GetVoicePreference implementa tts.PreferredVoiceLookup. 2 colunas fixas
// (preferred_voice_ja/en) em vez de interpolar o nome da coluna a partir de
// language — SELECT as duas sempre, escolhe em Go qual devolver.
func (r *PostgresRepository) GetVoicePreference(ctx context.Context, userID uuid.UUID, language string) (string, error) {
	row := r.pool.QueryRow(ctx, `SELECT preferred_voice_ja, preferred_voice_en FROM users WHERE id = $1`, userID)

	var ja, en *string
	if err := row.Scan(&ja, &en); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	var preferred *string
	switch primaryLanguageSubtag(language) {
	case "ja":
		preferred = ja
	case "en":
		preferred = en
	}
	if preferred == nil {
		return "", nil
	}
	return *preferred, nil
}

// SetVoicePreference grava a preferência de voz do usuário pro idioma
// pedido — voiceID="" limpa a preferência (volta pro default do idioma).
func (r *PostgresRepository) SetVoicePreference(ctx context.Context, userID uuid.UUID, language, voiceID string) error {
	var value *string
	if voiceID != "" {
		value = &voiceID
	}

	var query string
	switch primaryLanguageSubtag(language) {
	case "ja":
		query = `UPDATE users SET preferred_voice_ja = $2 WHERE id = $1`
	case "en":
		query = `UPDATE users SET preferred_voice_en = $2 WHERE id = $1`
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedVoiceLanguage, language)
	}

	tag, err := r.pool.Exec(ctx, query, userID, value)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdateAppearance aplica um patch parcial (só os campos não-nil de
// update são tocados) — monta o UPDATE dinamicamente em vez de sempre
// setar as 2 colunas, pra um PATCH só de tema não sobrescrever o accent
// salvo (e vice-versa).
func (r *PostgresRepository) UpdateAppearance(ctx context.Context, userID uuid.UUID, update AppearanceUpdate) error {
	var sets []string
	var args []any
	args = append(args, userID)

	if update.Theme != nil {
		var value *string
		if *update.Theme != "" {
			value = update.Theme
		}
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("theme = $%d", len(args)))
	}
	if update.AccentColor != nil {
		var value *string
		if *update.AccentColor != "" {
			value = update.AccentColor
		}
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("accent_color = $%d", len(args)))
	}
	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $1", strings.Join(sets, ", "))
	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func primaryLanguageSubtag(language string) string {
	return strings.ToLower(strings.SplitN(language, "-", 2)[0])
}
