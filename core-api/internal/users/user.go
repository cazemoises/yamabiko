package users

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID                uuid.UUID  `json:"id"`
	Email             string     `json:"email"`
	Name              string     `json:"name"`
	CreatedAt         time.Time  `json:"created_at"`
	CurrentSprintDay  int        `json:"current_sprint_day"`
	XPTotal           int        `json:"xp_total"`
	CurrentStreakDays int        `json:"current_streak_days"`
	LongestStreakDays int        `json:"longest_streak_days"`
	LastAttemptDate   *time.Time `json:"last_attempt_date,omitempty"`
	Badges            []string   `json:"badges"`
	PreferredVoiceJA  *string    `json:"preferred_voice_ja,omitempty"`
	PreferredVoiceEN  *string    `json:"preferred_voice_en,omitempty"`
}
