package exercises

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrExerciseNotFound = errors.New("exercício não encontrado")

type Filter struct {
	SprintDay  *int
	Category   *string
	Difficulty *int
}

type Repository interface {
	List(ctx context.Context, filter Filter) ([]Exercise, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Exercise, error)
}
