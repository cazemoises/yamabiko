package validation

import "encoding/json"

// MultipleChoiceTranslationData é o type_data de exercise_type
// "multiple_choice_translation" — Options no idioma-alvo (ou PT-BR,
// dependendo da direção do exercício), CorrectIndex aponta a opção certa.
type MultipleChoiceTranslationData struct {
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correct_index"`
}

func validateMultipleChoiceTranslation(typeData json.RawMessage, req AnswerRequest) (AnswerResult, error) {
	var data MultipleChoiceTranslationData
	if err := json.Unmarshal(typeData, &data); err != nil {
		return AnswerResult{}, ErrInvalidTypeData
	}
	return validateIndexAnswer(data.CorrectIndex, req)
}
