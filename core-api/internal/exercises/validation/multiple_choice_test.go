package validation_test

import (
	"encoding/json"
	"testing"

	"github.com/yamabiko/core-api/internal/exercises/validation"
)

func mcTypeData(t *testing.T, options []string, correctIndex int) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(validation.MultipleChoiceTranslationData{Options: options, CorrectIndex: correctIndex})
	if err != nil {
		t.Fatalf("falha ao montar type_data de teste: %v", err)
	}
	return raw
}

func TestValidateAnswer_MultipleChoiceTranslation_Correct(t *testing.T) {
	typeData := mcTypeData(t, []string{"good morning", "good night", "good afternoon"}, 0)
	idx := 0

	result, err := validation.ValidateAnswer("multiple_choice_translation", typeData, validation.AnswerRequest{SelectedIndex: &idx})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !result.Correct {
		t.Fatalf("esperava correct=true, veio false")
	}
	if result.CorrectIndex == nil || *result.CorrectIndex != 0 {
		t.Fatalf("esperava correct_index=0 no feedback, veio %v", result.CorrectIndex)
	}
}

func TestValidateAnswer_MultipleChoiceTranslation_Incorrect(t *testing.T) {
	typeData := mcTypeData(t, []string{"good morning", "good night", "good afternoon"}, 0)
	idx := 1

	result, err := validation.ValidateAnswer("multiple_choice_translation", typeData, validation.AnswerRequest{SelectedIndex: &idx})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Correct {
		t.Fatalf("esperava correct=false, veio true")
	}
}

// Edge case: selected_index fora do range das opções não pode dar panic nem
// erro — é só uma resposta errada como outra qualquer (o cliente pode mandar
// um índice inválido por bug de UI, isso não deveria derrubar o endpoint).
func TestValidateAnswer_MultipleChoiceTranslation_OutOfRangeIndexIsIncorrectNotError(t *testing.T) {
	typeData := mcTypeData(t, []string{"good morning", "good night"}, 0)
	idx := 99

	result, err := validation.ValidateAnswer("multiple_choice_translation", typeData, validation.AnswerRequest{SelectedIndex: &idx})
	if err != nil {
		t.Fatalf("esperava sem erro pra índice fora do range, veio %v", err)
	}
	if result.Correct {
		t.Fatalf("esperava correct=false pra índice fora do range, veio true")
	}
}

func TestValidateAnswer_MultipleChoiceTranslation_MissingSelectedIndex(t *testing.T) {
	typeData := mcTypeData(t, []string{"good morning", "good night"}, 0)

	_, err := validation.ValidateAnswer("multiple_choice_translation", typeData, validation.AnswerRequest{})
	if err != validation.ErrMissingAnswerField {
		t.Fatalf("esperava ErrMissingAnswerField, veio %v", err)
	}
}
