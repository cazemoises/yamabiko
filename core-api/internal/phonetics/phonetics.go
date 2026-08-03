// Package phonetics agrega os padrões de erro fonético recorrentes de um
// usuário (phonetic_error_patterns, Sec. 2 do CLAUDE.md), alimentado pelo
// phonetic_diff de cada attempt.
package phonetics

import (
	"context"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/comparison"
)

type PatternCount struct {
	Pattern     comparison.ErrorPattern `json:"pattern"`
	Occurrences int                     `json:"occurrences"`
}

type Repository interface {
	IncrementPatterns(ctx context.Context, userID uuid.UUID, counts map[comparison.ErrorPattern]int) error
	Heatmap(ctx context.Context, userID uuid.UUID) ([]PatternCount, error)
}

// CountPatterns conta quantas vezes cada padrão aparece no diff de uma tentativa.
func CountPatterns(diff []comparison.DiffEntry) map[comparison.ErrorPattern]int {
	counts := make(map[comparison.ErrorPattern]int)
	for _, entry := range diff {
		if entry.Pattern == "" {
			continue
		}
		counts[entry.Pattern]++
	}
	return counts
}
