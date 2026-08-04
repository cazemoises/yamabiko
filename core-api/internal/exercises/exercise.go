package exercises

import "github.com/google/uuid"

type Exercise struct {
	ID                 uuid.UUID `json:"id"`
	Category           string    `json:"category"`
	Difficulty         int       `json:"difficulty"`
	PromptPT           string    `json:"prompt_pt"`
	ExpectedTranscript string    `json:"expected_transcript"`
	ExpectedRomaji     string    `json:"expected_romaji,omitempty"`
	SprintDayRef       int       `json:"sprint_day_ref"`
	Language           string    `json:"language"`
}
