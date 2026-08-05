package validation

import "encoding/json"

// TrueFalseData é o type_data de exercise_type "true_false".
type TrueFalseData struct {
	Statement     string `json:"statement"`
	CorrectAnswer bool   `json:"correct_answer"`
}

func validateTrueFalse(typeData json.RawMessage, req AnswerRequest) (AnswerResult, error) {
	var data TrueFalseData
	if err := json.Unmarshal(typeData, &data); err != nil {
		return AnswerResult{}, ErrInvalidTypeData
	}
	if req.Answer == nil {
		return AnswerResult{}, ErrMissingAnswerField
	}

	correctAnswer := data.CorrectAnswer
	return AnswerResult{Correct: *req.Answer == data.CorrectAnswer, CorrectBool: &correctAnswer}, nil
}
