// Package validation valida respostas dos tipos de exercício sem áudio
// (Sec. pedida pelo usuário: multiple_choice_translation, word_order,
// verb_conjugation, dictation, free_translation, matching_pairs,
// true_false). Não toca no fluxo de áudio existente (attempts/comparison/
// stt-service) — dictation e free_translation reaproveitam comparison/
// diretamente (ver dictation.go/free_translation.go), os outros 5 são
// validação binária certo/errado contra o type_data do exercício.
package validation

import (
	"encoding/json"
	"errors"
)

// ErrInvalidTypeData é devolvido quando o type_data salvo no exercício não
// bate com a estrutura esperada pro exercise_type — indica dado mal
// cadastrado no banco, não erro do cliente.
var ErrInvalidTypeData = errors.New("validation: type_data inválido pro exercise_type")

// ErrMissingAnswerField é devolvido quando o payload do cliente não traz o
// campo de resposta que o exercise_type exige (ex: selected_index ausente
// pra multiple_choice_translation) — diferente de "resposta errada", é
// requisição malformada (o handler deve devolver 400, não processar como
// tentativa).
var ErrMissingAnswerField = errors.New("validation: campo de resposta ausente ou no formato errado pro exercise_type")

// ErrUnsupportedType é devolvido quando ValidateAnswer é chamado pra um
// exercise_type que não usa POST /answer (audio_pronunciation, dictation e
// free_translation têm fluxos próprios — os 2 últimos via POST
// /text-attempt).
var ErrUnsupportedType = errors.New("validation: exercise_type não suporta POST /exercises/{id}/answer")

// MatchingPair é compartilhado entre o type_data de matching_pairs
// (canônico) e o payload de resposta do cliente (o que ele montou) — mesma
// estrutura {left, right} nos dois lados.
type MatchingPair struct {
	Left  string `json:"left"`
	Right string `json:"right"`
}

// AnswerRequest é o corpo de POST /exercises/{id}/answer — só o(s) campo(s)
// relevante(s) pro exercise_type do exercício precisam vir preenchidos, os
// outros ficam nil/vazios e são ignorados pelo validador daquele tipo.
type AnswerRequest struct {
	SelectedIndex  *int           `json:"selected_index,omitempty"`
	SubmittedOrder []string       `json:"submitted_order,omitempty"`
	SubmittedPairs []MatchingPair `json:"submitted_pairs,omitempty"`
	Answer         *bool          `json:"answer,omitempty"`
}

// AnswerResult é a resposta de POST /exercises/{id}/answer — Correct é o
// veredito binário pedido pelo usuário; os campos Correct* carregam a
// resposta canônica no formato do próprio tipo, pro cliente mostrar
// feedback ("a resposta certa era X") sem reimplementar o parsing de
// type_data. Só o campo relevante ao exercise_type vem preenchido.
type AnswerResult struct {
	Correct      bool           `json:"correct"`
	CorrectIndex *int           `json:"correct_index,omitempty"`
	CorrectOrder []string       `json:"correct_order,omitempty"`
	CorrectPairs []MatchingPair `json:"correct_pairs,omitempty"`
	CorrectBool  *bool          `json:"correct_answer,omitempty"`
}

// ValidateAnswer despacha pro validador do exercise_type indicado.
func ValidateAnswer(exerciseType string, typeData json.RawMessage, req AnswerRequest) (AnswerResult, error) {
	switch exerciseType {
	case "multiple_choice_translation":
		return validateMultipleChoiceTranslation(typeData, req)
	case "word_order":
		return validateWordOrder(typeData, req)
	case "verb_conjugation":
		return validateVerbConjugation(typeData, req)
	case "matching_pairs":
		return validateMatchingPairs(typeData, req)
	case "true_false":
		return validateTrueFalse(typeData, req)
	default:
		return AnswerResult{}, ErrUnsupportedType
	}
}

// validateIndexAnswer é o núcleo compartilhado por multiple_choice_translation
// e verb_conjugation — os dois são "escolha entre opções, uma delas certa",
// só o type_data em volta muda.
func validateIndexAnswer(correctIndex int, req AnswerRequest) (AnswerResult, error) {
	if req.SelectedIndex == nil {
		return AnswerResult{}, ErrMissingAnswerField
	}
	idx := correctIndex
	return AnswerResult{Correct: *req.SelectedIndex == correctIndex, CorrectIndex: &idx}, nil
}
