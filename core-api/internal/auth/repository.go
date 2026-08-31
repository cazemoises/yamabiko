package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("usuário não encontrado")

// UserRepository resolve a identidade confiada pelos headers do Pangolin
// (ver context.go) pro usuário correspondente no yamabiko — sem senha nem
// PIN, o Pangolin já fez a autenticação antes da requisição chegar aqui.
type UserRepository interface {
	// FindOrCreateByEmail devolve o ID do usuário com esse email, criando-o
	// na hora (com `name`) se for o primeiro acesso dessa pessoa.
	FindOrCreateByEmail(ctx context.Context, email, name string) (uuid.UUID, error)
}
