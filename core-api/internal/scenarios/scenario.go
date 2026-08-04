package scenarios

import "github.com/google/uuid"

// Scenario agrupa exercícios numa sequência narrativa única (ex: "chegar no
// escritório de manhã") em vez de exercícios soltos — Sec. pedida pelo
// usuário. Não substitui exercícios soltos: um exercício sem scenario_id
// continua funcionando exatamente como antes (Filter.ScenarioID nil).
type Scenario struct {
	ID                   uuid.UUID `json:"id"`
	Language             string    `json:"language"`
	TitlePT              string    `json:"title_pt"`
	ContextDescriptionPT string    `json:"context_description_pt"`
	OrderIndex           int       `json:"order_index"`
}
