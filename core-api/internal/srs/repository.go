package srs

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Get devolve o card salvo pra (userID, chunkID), ou um Card zero com
	// StatusNovo se essa dupla nunca foi revisada.
	Get(ctx context.Context, userID, chunkID uuid.UUID) (Card, Status, error)
	Save(ctx context.Context, userID, chunkID uuid.UUID, review Review) error
	CountByStatus(ctx context.Context, userID uuid.UUID) (map[Status]int, error)
}
