package validation

import "encoding/json"

// WordOrderData é o type_data de exercise_type "word_order" —
// ShuffledWords é o que o cliente mostra pro aluno reordenar, CorrectOrder é
// o gabarito na ordem certa.
type WordOrderData struct {
	ShuffledWords []string `json:"shuffled_words"`
	CorrectOrder  []string `json:"correct_order"`
}

func validateWordOrder(typeData json.RawMessage, req AnswerRequest) (AnswerResult, error) {
	var data WordOrderData
	if err := json.Unmarshal(typeData, &data); err != nil {
		return AnswerResult{}, ErrInvalidTypeData
	}

	correct := len(req.SubmittedOrder) == len(data.CorrectOrder)
	if correct {
		for i, word := range req.SubmittedOrder {
			if word != data.CorrectOrder[i] {
				correct = false
				break
			}
		}
	}

	return AnswerResult{Correct: correct, CorrectOrder: data.CorrectOrder}, nil
}
