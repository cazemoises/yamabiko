package exercises

import (
	"encoding/json"

	"github.com/google/uuid"
)

// ExerciseType enumera os 8 formatos suportados — "audio_pronunciation" é o
// original (grava áudio, comparado via stt-service + comparison/), os outros
// 7 são os tipos de exercício sem áudio (Sec. pedida pelo usuário),
// validados em internal/exercises/validation/. Mesmos valores do CHECK
// constraint da migration 0016 — mantidos como string (não um tipo Go
// dedicado) porque é isso que trafega no JSON e na coluna TEXT do Postgres.
type ExerciseType = string

const (
	ExerciseTypeAudioPronunciation        ExerciseType = "audio_pronunciation"
	ExerciseTypeMultipleChoiceTranslation ExerciseType = "multiple_choice_translation"
	ExerciseTypeWordOrder                 ExerciseType = "word_order"
	ExerciseTypeVerbConjugation           ExerciseType = "verb_conjugation"
	ExerciseTypeDictation                 ExerciseType = "dictation"
	ExerciseTypeFreeTranslation           ExerciseType = "free_translation"
	ExerciseTypeMatchingPairs             ExerciseType = "matching_pairs"
	ExerciseTypeTrueFalse                 ExerciseType = "true_false"
)

type Exercise struct {
	ID                 uuid.UUID       `json:"id"`
	Category           string          `json:"category"`
	Difficulty         int             `json:"difficulty"`
	PromptPT           string          `json:"prompt_pt"`
	ExpectedTranscript string          `json:"expected_transcript"`
	ExpectedRomaji     string          `json:"expected_romaji,omitempty"`
	SprintDayRef       int             `json:"sprint_day_ref"`
	Language           string          `json:"language"`
	ScenarioID         *uuid.UUID      `json:"scenario_id,omitempty"`
	OrderInScenario    *int            `json:"order_in_scenario,omitempty"`
	ExerciseType       ExerciseType    `json:"exercise_type"`
	TypeData           json.RawMessage `json:"type_data,omitempty"`
}
