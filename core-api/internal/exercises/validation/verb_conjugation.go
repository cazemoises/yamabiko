package validation

import "encoding/json"

// VerbConjugationData é o type_data de exercise_type "verb_conjugation" —
// SentenceTemplate/VerbInfinitive são só pra exibição (o "___" na frase e o
// infinitivo do verbo sendo conjugado), a validação em si é a mesma
// escolha-por-índice de multiple_choice_translation.
type VerbConjugationData struct {
	SentenceTemplate string   `json:"sentence_template"`
	VerbInfinitive   string   `json:"verb_infinitive"`
	Options          []string `json:"options"`
	CorrectIndex     int      `json:"correct_index"`
}

func validateVerbConjugation(typeData json.RawMessage, req AnswerRequest) (AnswerResult, error) {
	var data VerbConjugationData
	if err := json.Unmarshal(typeData, &data); err != nil {
		return AnswerResult{}, ErrInvalidTypeData
	}
	return validateIndexAnswer(data.CorrectIndex, req)
}
