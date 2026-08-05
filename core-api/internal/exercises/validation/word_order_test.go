package validation_test

import (
	"encoding/json"
	"testing"

	"github.com/yamabiko/core-api/internal/exercises/validation"
)

func woTypeData(t *testing.T, shuffled, correct []string) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(validation.WordOrderData{ShuffledWords: shuffled, CorrectOrder: correct})
	if err != nil {
		t.Fatalf("falha ao montar type_data de teste: %v", err)
	}
	return raw
}

func TestValidateAnswer_WordOrder_Correct(t *testing.T) {
	typeData := woTypeData(t, []string{"は", "私", "です", "学生"}, []string{"私", "は", "学生", "です"})

	result, err := validation.ValidateAnswer("word_order", typeData, validation.AnswerRequest{
		SubmittedOrder: []string{"私", "は", "学生", "です"},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !result.Correct {
		t.Fatalf("esperava correct=true, veio false")
	}
	if len(result.CorrectOrder) != 4 {
		t.Fatalf("esperava correct_order no feedback com 4 palavras, veio %v", result.CorrectOrder)
	}
}

func TestValidateAnswer_WordOrder_Incorrect(t *testing.T) {
	typeData := woTypeData(t, []string{"は", "私", "です", "学生"}, []string{"私", "は", "学生", "です"})

	result, err := validation.ValidateAnswer("word_order", typeData, validation.AnswerRequest{
		SubmittedOrder: []string{"は", "私", "です", "学生"},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Correct {
		t.Fatalf("esperava correct=false, veio true")
	}
}

// Edge case: ordem submetida com tamanho diferente do gabarito (palavra
// faltando ou duplicada) precisa dar incorreto sem panic de index-out-of-range.
func TestValidateAnswer_WordOrder_DifferentLengthIsIncorrectNotPanic(t *testing.T) {
	typeData := woTypeData(t, []string{"は", "私", "です", "学生"}, []string{"私", "は", "学生", "です"})

	result, err := validation.ValidateAnswer("word_order", typeData, validation.AnswerRequest{
		SubmittedOrder: []string{"私", "は", "学生"},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Correct {
		t.Fatalf("esperava correct=false pra submissão com tamanho diferente, veio true")
	}
}
