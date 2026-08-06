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
// Devolve também QUAL acceptable_answer venceu (matchedAnswer) — o
// frontend precisa saber isso pra desenhar o diff "Esperado x Você
// digitou" (Frame 18/19), já que type_data tem várias respostas possíveis
// e só o backend sabe contra qual delas o diff foi calculado.
func ValidateFreeTranslation(typeData json.RawMessage, language, submittedTranslation string) (result comparison.Result, matchedAnswer string, err error) {
	var data FreeTranslationData
	if err := json.Unmarshal(typeData, &data); err != nil {
		return comparison.Result{}, "", ErrInvalidTypeData
	}
	if len(data.AcceptableAnswers) == 0 {
		return comparison.Result{}, "", ErrInvalidTypeData
	}

	bestAnswer := data.AcceptableAnswers[0]
	best := comparison.CompareLang(bestAnswer, submittedTranslation, language)
	for _, answer := range data.AcceptableAnswers[1:] {
		if r := comparison.CompareLang(answer, submittedTranslation, language); r.SimilarityScore > best.SimilarityScore {
			best = r
			bestAnswer = answer
		}
	}
	return best, bestAnswer, nil
}
