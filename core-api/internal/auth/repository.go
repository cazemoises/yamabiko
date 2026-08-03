package auth

import (
	"context"
	"errors"

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
}
