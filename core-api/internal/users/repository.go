package users

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/gamification"
)

var ErrUserNotFound = errors.New("usuário não encontrado")

type Repository interface {
	FindProfileByID(ctx context.Context, id uuid.UUID) (*Profile, error)
	FindStatsByID(ctx context.Context, id uuid.UUID) (*gamification.UserStats, error)
	UpdateStats(ctx context.Context, userID uuid.UUID, stats gamification.UserStats) error
	EarnedBadges(ctx context.Context, userID uuid.UUID) (map[gamification.Badge]bool, error)
	AwardBadges(ctx context.Context, userID uuid.UUID, badges []gamification.Badge) error
}
