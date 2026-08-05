package validation_test

import (
	"encoding/json"
	"testing"

	"github.com/yamabiko/core-api/internal/exercises/validation"
)

func mpTypeData(t *testing.T, pairs []validation.MatchingPair) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(validation.MatchingPairsData{Pairs: pairs})
	if err != nil {
		t.Fatalf("falha ao montar type_data de teste: %v", err)
	}
	return raw
}

var canonicalPairs = []validation.MatchingPair{
	{Left: "bom dia", Right: "おはよう"},
	{Left: "obrigado", Right: "ありがとう"},
	{Left: "desculpa", Right: "すみません"},
}

func TestValidateAnswer_MatchingPairs_Correct(t *testing.T) {
	typeData := mpTypeData(t, canonicalPairs)

	result, err := validation.ValidateAnswer("matching_pairs", typeData, validation.AnswerRequest{
		SubmittedPairs: []validation.MatchingPair{
			{Left: "obrigado", Right: "ありがとう"},
			{Left: "bom dia", Right: "おはよう"},
			{Left: "desculpa", Right: "すみません"},
		},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !result.Correct {
		t.Fatalf("esperava correct=true pra todos os pares certos (ordem diferente da canônica), veio false")
	}
}

func TestValidateAnswer_MatchingPairs_Incorrect(t *testing.T) {
	typeData := mpTypeData(t, canonicalPairs)

	result, err := validation.ValidateAnswer("matching_pairs", typeData, validation.AnswerRequest{
		SubmittedPairs: []validation.MatchingPair{
			{Left: "bom dia", Right: "ありがとう"},
			{Left: "obrigado", Right: "おはよう"},
			{Left: "desculpa", Right: "すみません"},
		},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Correct {
		t.Fatalf("esperava correct=false pra pares trocados, veio true")
	}
}

// Edge case: validação é binária (Sec. pedida pelo usuário) — acertar 2 de
// 3 pares NÃO conta como correto, é tudo ou nada.
func TestValidateAnswer_MatchingPairs_PartiallyCorrectCountsAsIncorrect(t *testing.T) {
	typeData := mpTypeData(t, canonicalPairs)

	result, err := validation.ValidateAnswer("matching_pairs", typeData, validation.AnswerRequest{
		SubmittedPairs: []validation.MatchingPair{
			{Left: "bom dia", Right: "おはよう"},
			{Left: "obrigado", Right: "ありがとう"},
			{Left: "desculpa", Right: "ありがとう"}, // errado de propósito
		},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Correct {
		t.Fatalf("esperava correct=false pra 2/3 pares certos (binário, sem crédito parcial), veio true")
	}
}
