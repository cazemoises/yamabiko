package attempts

import (
	"time"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/comparison"
)

type Attempt struct {
	ID              uuid.UUID              `json:"id"`
	UserID          uuid.UUID              `json:"user_id"`
	ExerciseID      uuid.UUID              `json:"exercise_id"`
	AudioTranscript string                 `json:"audio_transcript"`
	SimilarityScore float64                `json:"similarity_score"`
	Verdict         comparison.Verdict     `json:"verdict"`
	PhoneticDiff    []comparison.DiffEntry `json:"phonetic_diff"`
	CreatedAt       time.Time              `json:"created_at"`
}
