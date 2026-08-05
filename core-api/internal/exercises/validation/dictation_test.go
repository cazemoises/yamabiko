package validation_test

import (
	"testing"

	"github.com/yamabiko/core-api/internal/comparison"
	"github.com/yamabiko/core-api/internal/exercises/validation"
)

func TestValidateDictation_Correct(t *testing.T) {
	result := validation.ValidateDictation("おはようございます", "ja-JP", "おはようございます")

	if result.Verdict != comparison.VerdictPass {
		t.Fatalf("esperava PASS pra ditado idêntico, veio %s (score %.2f)", result.Verdict, result.SimilarityScore)
	}
}

func TestValidateDictation_Incorrect(t *testing.T) {
	result := validation.ValidateDictation("おはようございます", "ja-JP", "こんばんは")

	if result.Verdict == comparison.VerdictPass {
		t.Fatalf("esperava PARTIAL/FAIL pra ditado bem diferente, veio PASS (score %.2f)", result.SimilarityScore)
	}
}

// Edge case: transcript submetido vazio (aluno não digitou nada) não pode
// dar erro nem panic — é só o score mais baixo possível, FAIL normal.
func TestValidateDictation_EmptySubmission(t *testing.T) {
	result := validation.ValidateDictation("おはようございます", "ja-JP", "")

	if result.Verdict != comparison.VerdictFail {
		t.Fatalf("esperava FAIL pra submissão vazia, veio %s", result.Verdict)
	}
	if result.SimilarityScore != 0 {
		t.Fatalf("esperava score=0 pra submissão vazia, veio %.2f", result.SimilarityScore)
	}
}
