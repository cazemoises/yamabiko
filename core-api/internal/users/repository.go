package users

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/gamification"
)

var ErrUserNotFound = errors.New("usuário não encontrado")

// ErrUnsupportedVoiceLanguage é devolvido por SetVoicePreference quando o
// idioma pedido não tem coluna de preferência (só ja-JP/en-US existem hoje)
// — erro de validação do request, não falha de infra.
var ErrUnsupportedVoiceLanguage = errors.New("users: idioma sem preferência de voz suportada")

type Repository interface {
	FindProfileByID(ctx context.Context, id uuid.UUID) (*Profile, error)
	FindStatsByID(ctx context.Context, id uuid.UUID) (*gamification.UserStats, error)
	UpdateStats(ctx context.Context, userID uuid.UUID, stats gamification.UserStats) error
	EarnedBadges(ctx context.Context, userID uuid.UUID) (map[gamification.Badge]bool, error)
	AwardBadges(ctx context.Context, userID uuid.UUID, badges []gamification.Badge) error
	// GetVoicePreference/SetVoicePreference implementam tts.PreferredVoiceLookup
	// (satisfação estrutural, sem o pacote users importar tts — ver
	// core-api/internal/tts/service.go).
	GetVoicePreference(ctx context.Context, userID uuid.UUID, language string) (voiceID string, err error)
	SetVoicePreference(ctx context.Context, userID uuid.UUID, language, voiceID string) error
}
