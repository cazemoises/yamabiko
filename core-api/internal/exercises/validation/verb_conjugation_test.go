package validation_test

import (
	"encoding/json"
	"testing"

	"github.com/yamabiko/core-api/internal/exercises/validation"
)

func vcTypeData(t *testing.T, template, infinitive string, options []string, correctIndex int) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(validation.VerbConjugationData{
		SentenceTemplate: template,
		VerbInfinitive:   infinitive,
		Options:          options,
		CorrectIndex:     correctIndex,
	})
	if err != nil {
		t.Fatalf("falha ao montar type_data de teste: %v", err)
	}
	return raw
}

func TestValidateAnswer_VerbConjugation_Correct(t *testing.T) {
	typeData := vcTypeData(t, "私は毎日パンを___。", "食べる", []string{"食べます", "食べました", "食べません"}, 0)
	idx := 0

	result, err := validation.ValidateAnswer("verb_conjugation", typeData, validation.AnswerRequest{SelectedIndex: &idx})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !result.Correct {
		t.Fatalf("esperava correct=true, veio false")
	}
}

func TestValidateAnswer_VerbConjugation_Incorrect(t *testing.T) {
	typeData := vcTypeData(t, "私は毎日パンを___。", "食べる", []string{"食べます", "食べました", "食べません"}, 0)
	idx := 2

	result, err := validation.ValidateAnswer("verb_conjugation", typeData, validation.AnswerRequest{SelectedIndex: &idx})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Correct {
		t.Fatalf("esperava correct=false, veio true")
	}
}

// Edge case: campo selected_index ausente é requisição malformada
// (ErrMissingAnswerField), não deve ser tratado como "escolheu o índice 0"
// silenciosamente — o zero-value de *int não pode se confundir com uma
// escolha real do usuário.
func TestValidateAnswer_VerbConjugation_MissingSelectedIndex(t *testing.T) {
	typeData := vcTypeData(t, "私は毎日パンを___。", "食べる", []string{"食べます", "食べました", "食べません"}, 0)

	_, err := validation.ValidateAnswer("verb_conjugation", typeData, validation.AnswerRequest{})
	if err != validation.ErrMissingAnswerField {
		t.Fatalf("esperava ErrMissingAnswerField, veio %v", err)
	}
}
