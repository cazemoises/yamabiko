package validation_test

import (
	"encoding/json"
	"testing"

	"github.com/yamabiko/core-api/internal/comparison"
	"github.com/yamabiko/core-api/internal/exercises/validation"
)

func ftTypeData(t *testing.T, acceptable []string) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(validation.FreeTranslationData{AcceptableAnswers: acceptable})
	if err != nil {
		t.Fatalf("falha ao montar type_data de teste: %v", err)
	}
	return raw
}

func TestValidateFreeTranslation_Correct(t *testing.T) {
	typeData := ftTypeData(t, []string{"I am a student", "I'm a student"})

	result, err := validation.ValidateFreeTranslation(typeData, "en-US", "I am a student")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Verdict != comparison.VerdictPass {
		t.Fatalf("esperava PASS pra tradução idêntica a uma acceptable_answer, veio %s", result.Verdict)
	}
}

func TestValidateFreeTranslation_Incorrect(t *testing.T) {
	typeData := ftTypeData(t, []string{"I am a student", "I'm a student"})

	result, err := validation.ValidateFreeTranslation(typeData, "en-US", "the weather is nice today")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Verdict == comparison.VerdictPass {
		t.Fatalf("esperava PARTIAL/FAIL pra tradução sem relação com nenhuma acceptable_answer, veio PASS")
	}
}

// Edge case: com várias respostas aceitáveis, o resultado tem que ser o
// MELHOR score entre todas — não o da primeira da lista. Aqui a submissão
// bate exatamente com a 2ª opção, não a 1ª.
func TestValidateFreeTranslation_PicksBestScoreNotFirst(t *testing.T) {
	typeData := ftTypeData(t, []string{"the weather is nice today", "I'm a student"})

	result, err := validation.ValidateFreeTranslation(typeData, "en-US", "I'm a student")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Verdict != comparison.VerdictPass {
		t.Fatalf("esperava PASS (melhor score é a 2ª acceptable_answer, idêntica), veio %s (score %.2f)", result.Verdict, result.SimilarityScore)
	}
}
