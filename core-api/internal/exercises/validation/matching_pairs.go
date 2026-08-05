package validation

import "encoding/json"

// MatchingPairsData é o type_data de exercise_type "matching_pairs" — Left
// em português, Right no idioma-alvo (Sec. pedida pelo usuário).
type MatchingPairsData struct {
	Pairs []MatchingPair `json:"pairs"`
}

// validateMatchingPairs é binário (Sec. pedida pelo usuário): correct só
// quando TODOS os pares submetidos batem com o canônico, sem crédito
// parcial. A comparação ignora a ordem em que o cliente submeteu os pares
// (o aluno pode ter arrastado na ordem que quis) — usa left como chave pra
// achar o right esperado.
func validateMatchingPairs(typeData json.RawMessage, req AnswerRequest) (AnswerResult, error) {
	var data MatchingPairsData
	if err := json.Unmarshal(typeData, &data); err != nil {
		return AnswerResult{}, ErrInvalidTypeData
	}

	canonical := make(map[string]string, len(data.Pairs))
	for _, p := range data.Pairs {
		canonical[p.Left] = p.Right
	}

	correct := len(req.SubmittedPairs) == len(data.Pairs)
	if correct {
		for _, p := range req.SubmittedPairs {
			if canonical[p.Left] != p.Right {
				correct = false
				break
			}
		}
	}

	return AnswerResult{Correct: correct, CorrectPairs: data.Pairs}, nil
}
