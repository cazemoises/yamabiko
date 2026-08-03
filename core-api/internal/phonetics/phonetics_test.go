package phonetics_test

import (
	"testing"

	"github.com/yamabiko/core-api/internal/comparison"
	"github.com/yamabiko/core-api/internal/phonetics"
)

func TestCountPatterns_CountsRecognizedPatternsAndIgnoresMatches(t *testing.T) {
	diff := []comparison.DiffEntry{
		{Op: comparison.OpDelete, Pattern: comparison.PatternHAspiradoOmitido},
		{Op: comparison.OpDelete, Pattern: comparison.PatternHAspiradoOmitido},
		{Op: comparison.OpSubstitute, Pattern: comparison.PatternRLTConfusao},
		{Op: comparison.OpInsert, Pattern: comparison.PatternOutro},
	}

	counts := phonetics.CountPatterns(diff)

	if counts[comparison.PatternHAspiradoOmitido] != 2 {
		t.Fatalf("esperava 2 ocorrências de H_ASPIRADO_OMITIDO, veio %d", counts[comparison.PatternHAspiradoOmitido])
	}
	if counts[comparison.PatternRLTConfusao] != 1 {
		t.Fatalf("esperava 1 ocorrência de R_L_T_CONFUSAO, veio %d", counts[comparison.PatternRLTConfusao])
	}
	if counts[comparison.PatternOutro] != 1 {
		t.Fatalf("esperava 1 ocorrência de OUTRO, veio %d", counts[comparison.PatternOutro])
	}
}

func TestCountPatterns_EmptyDiff_ReturnsEmptyMap(t *testing.T) {
	counts := phonetics.CountPatterns(nil)
	if len(counts) != 0 {
		t.Fatalf("esperava mapa vazio, veio %v", counts)
	}
}
