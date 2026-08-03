package attempts

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, attempt *Attempt) error
	ListByUserAndExercise(ctx context.Context, userID, exerciseID uuid.UUID) ([]Attempt, error)
}
