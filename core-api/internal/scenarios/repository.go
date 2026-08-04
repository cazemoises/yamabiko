package scenarios

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrScenarioNotFound = errors.New("cenário não encontrado")

type Filter struct {
	Language *string
}

type Repository interface {
	List(ctx context.Context, filter Filter) ([]Scenario, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Scenario, error)
}
