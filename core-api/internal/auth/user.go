package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID
	Email            string
	PasswordHash     string
	Name             string
	CreatedAt        time.Time
	CurrentSprintDay int
	// AccentColor é a mesma coluna usada pra tema da UI (migration 0017,
	// internal/users) — reaproveitada aqui como o "campo equivalente já
	// usado no design do frontend" pra diferenciar perfis visualmente na
	// tela de seleção por PIN, sem duplicar o conceito.
	AccentColor *string
	// PinHash NULL = usuário não configurou login por PIN ainda (não
	// aparece em ListPinEnabledProfiles). PinFailedAttempts/PinLockedUntil
	// implementam lockout de 5 tentativas/15min — ver pin.go.
	PinHash           *string
	PinFailedAttempts int
	PinLockedUntil    *time.Time
}
