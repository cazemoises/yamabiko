package validation

import "github.com/yamabiko/core-api/internal/comparison"

// ValidateDictation é o exercise_type "dictation" — sem type_data (Sec.
// pedida pelo usuário), reaproveita expected_transcript do próprio
// exercício e a mesma engine de comparação (Levenshtein) usada pelo fluxo
// de áudio, só que texto-a-texto: o aluno digita o que ouviu em vez de
// gravar, sem passar pelo stt-service.
func ValidateDictation(expectedTranscript, language, submittedTranscript string) comparison.Result {
	return comparison.CompareLang(expectedTranscript, submittedTranscript, language)
}
