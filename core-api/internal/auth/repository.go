package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("usuário não encontrado")
	ErrEmailTaken   = errors.New("email já cadastrado")
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	// ListPinEnabledProfiles alimenta GET /auth/profiles — só usuários com
	// pin_hash configurado, sem nenhum dado sensível na projeção.
	ListPinEnabledProfiles(ctx context.Context) ([]PinProfile, error)
	UpdatePinAuthState(ctx context.Context, userID uuid.UUID, failedAttempts int, lockedUntil *time.Time) error
	SetPinHash(ctx context.Context, userID uuid.UUID, hash string) error
}
