package validation

import (
	"encoding/json"

	"github.com/yamabiko/core-api/internal/comparison"
)

// FreeTranslationData é o type_data de exercise_type "free_translation" —
// mais de uma tradução conta como certa (ex: "I am a student"/"I'm a
// student"), então a validação testa contra cada uma e fica com a melhor.
type FreeTranslationData struct {
	AcceptableAnswers []string `json:"acceptable_answers"`
}

// ValidateFreeTranslation compara submittedTranslation contra CADA
// acceptable_answers via comparison.CompareLang (mesma engine Levenshtein
// texto-a-texto de ValidateDictation) e devolve o melhor score — o aluno só
// precisa acertar uma das traduções aceitas, não a "primeira" da lista.
func ValidateFreeTranslation(typeData json.RawMessage, language, submittedTranslation string) (comparison.Result, error) {
	var data FreeTranslationData
	if err := json.Unmarshal(typeData, &data); err != nil {
		return comparison.Result{}, ErrInvalidTypeData
	}
	if len(data.AcceptableAnswers) == 0 {
		return comparison.Result{}, ErrInvalidTypeData
	}

	best := comparison.CompareLang(data.AcceptableAnswers[0], submittedTranslation, language)
	for _, answer := range data.AcceptableAnswers[1:] {
		if r := comparison.CompareLang(answer, submittedTranslation, language); r.SimilarityScore > best.SimilarityScore {
			best = r
		}
	}
	return best, nil
}
