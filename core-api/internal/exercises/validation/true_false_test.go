package validation_test

import (
	"encoding/json"
	"testing"

	"github.com/yamabiko/core-api/internal/exercises/validation"
)

func tfTypeData(t *testing.T, statement string, correctAnswer bool) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(validation.TrueFalseData{Statement: statement, CorrectAnswer: correctAnswer})
	if err != nil {
		t.Fatalf("falha ao montar type_data de teste: %v", err)
	}
	return raw
}

func TestValidateAnswer_TrueFalse_Correct(t *testing.T) {
	typeData := tfTypeData(t, "「ありがとう」significa \"obrigado\"", true)
	answer := true

	result, err := validation.ValidateAnswer("true_false", typeData, validation.AnswerRequest{Answer: &answer})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !result.Correct {
		t.Fatalf("esperava correct=true, veio false")
	}
	if result.CorrectBool == nil || *result.CorrectBool != true {
		t.Fatalf("esperava correct_answer=true no feedback, veio %v", result.CorrectBool)
	}
}

func TestValidateAnswer_TrueFalse_Incorrect(t *testing.T) {
	typeData := tfTypeData(t, "「さようなら」significa \"bom dia\"", false)
	answer := true

	result, err := validation.ValidateAnswer("true_false", typeData, validation.AnswerRequest{Answer: &answer})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Correct {
		t.Fatalf("esperava correct=false, veio true")
	}
}

// Edge case: campo answer ausente é requisição malformada
// (ErrMissingAnswerField) — o zero-value de *bool não pode se confundir com
// "o aluno respondeu false".
func TestValidateAnswer_TrueFalse_MissingAnswer(t *testing.T) {
	typeData := tfTypeData(t, "「ありがとう」significa \"obrigado\"", true)

	_, err := validation.ValidateAnswer("true_false", typeData, validation.AnswerRequest{})
	if err != validation.ErrMissingAnswerField {
		t.Fatalf("esperava ErrMissingAnswerField, veio %v", err)
	}
}
