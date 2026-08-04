package attempts

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/comparison"
	"github.com/yamabiko/core-api/internal/exercises"
	"github.com/yamabiko/core-api/internal/gamification"
	"github.com/yamabiko/core-api/internal/phonetics"
	"github.com/yamabiko/core-api/internal/srs"
	"github.com/yamabiko/core-api/internal/sttclient"
)

type Transcriber interface {
	Transcribe(ctx context.Context, filename string, audio io.Reader, language string) (*sttclient.TranscriptionResult, error)
}

type ExerciseFinder interface {
	FindByID(ctx context.Context, id uuid.UUID) (*exercises.Exercise, error)
}

// SRSRepository é a fatia de srs.Repository usada por attempts — cada exercício é
// tratado como um "chunk" pra fins de spaced repetition (ver migration 0006).
type SRSRepository interface {
	Get(ctx context.Context, userID, chunkID uuid.UUID) (srs.Card, srs.Status, error)
	Save(ctx context.Context, userID, chunkID uuid.UUID, review srs.Review) error
}

// UserStatsRepository é a fatia de users.Repository usada por attempts pra
// atualizar XP/streak/badges a cada tentativa.
type UserStatsRepository interface {
	FindStatsByID(ctx context.Context, userID uuid.UUID) (*gamification.UserStats, error)
	UpdateStats(ctx context.Context, userID uuid.UUID, stats gamification.UserStats) error
	EarnedBadges(ctx context.Context, userID uuid.UUID) (map[gamification.Badge]bool, error)
	AwardBadges(ctx context.Context, userID uuid.UUID, badges []gamification.Badge) error
}

type PhoneticsRepository interface {
	IncrementPatterns(ctx context.Context, userID uuid.UUID, language string, counts map[comparison.ErrorPattern]int) error
}

type Service struct {
	repo           Repository
	transcriber    Transcriber
	exerciseFinder ExerciseFinder
	srsRepo        SRSRepository
	userStats      UserStatsRepository
	phoneticsRepo  PhoneticsRepository
	now            func() time.Time
}

func NewService(
	repo Repository,
	transcriber Transcriber,
	exerciseFinder ExerciseFinder,
	srsRepo SRSRepository,
	userStats UserStatsRepository,
	phoneticsRepo PhoneticsRepository,
) *Service {
	return &Service{
		repo:           repo,
		transcriber:    transcriber,
		exerciseFinder: exerciseFinder,
		srsRepo:        srsRepo,
		userStats:      userStats,
		phoneticsRepo:  phoneticsRepo,
		now:            time.Now,
	}
}

type SubmitResult struct {
	Transcript string
	Score      float64
	Verdict    comparison.Verdict
	Diff       []comparison.DiffEntry
	XPGained   int
	NewBadges  []gamification.Badge
}

func (s *Service) Submit(ctx context.Context, userID, exerciseID uuid.UUID, filename string, audio io.Reader) (*SubmitResult, error) {
	exercise, err := s.exerciseFinder.FindByID(ctx, exerciseID)
	if err != nil {
		return nil, err
	}

	transcription, err := s.transcriber.Transcribe(ctx, filename, audio, exercise.Language)
	if err != nil {
		return nil, err
	}

	result := comparison.CompareLang(exercise.ExpectedTranscript, transcription.Transcript, exercise.Language)
	now := s.now()

	attempt := &Attempt{
		ID:              uuid.New(),
		UserID:          userID,
		ExerciseID:      exerciseID,
		AudioTranscript: transcription.Transcript,
		SimilarityScore: result.SimilarityScore,
		Verdict:         result.Verdict,
		PhoneticDiff:    result.PhoneticDiff,
		Language:        exercise.Language,
	}
	if err := s.repo.Create(ctx, attempt); err != nil {
		return nil, err
	}

	if err := s.reviewChunk(ctx, userID, exerciseID, result.SimilarityScore, now); err != nil {
		return nil, err
	}

	gamResult, err := s.updateGamification(ctx, userID, result.Verdict, now)
	if err != nil {
		return nil, err
	}

	if counts := phonetics.CountPatterns(result.PhoneticDiff); len(counts) > 0 {
		if err := s.phoneticsRepo.IncrementPatterns(ctx, userID, exercise.Language, counts); err != nil {
			return nil, err
		}
	}

	return &SubmitResult{
		Transcript: transcription.Transcript,
		Score:      result.SimilarityScore,
		Verdict:    result.Verdict,
		Diff:       result.PhoneticDiff,
		XPGained:   gamResult.XPGained,
		NewBadges:  gamResult.NewBadges,
	}, nil
}

func (s *Service) reviewChunk(ctx context.Context, userID, chunkID uuid.UUID, score float64, now time.Time) error {
	card, _, err := s.srsRepo.Get(ctx, userID, chunkID)
	if err != nil {
		return err
	}
	quality := srs.QualityFromScore(score)
	review := srs.Schedule(card, quality, now)
	return s.srsRepo.Save(ctx, userID, chunkID, review)
}

func (s *Service) updateGamification(ctx context.Context, userID uuid.UUID, verdict comparison.Verdict, now time.Time) (*gamification.UpdateResult, error) {
	stats, err := s.userStats.FindStatsByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	earnedBadges, err := s.userStats.EarnedBadges(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := gamification.RecordAttempt(*stats, verdict, now, earnedBadges)

	if err := s.userStats.UpdateStats(ctx, userID, result.Stats); err != nil {
		return nil, err
	}
	if len(result.NewBadges) > 0 {
		if err := s.userStats.AwardBadges(ctx, userID, result.NewBadges); err != nil {
			return nil, err
		}
	}

	return &result, nil
}

func (s *Service) History(ctx context.Context, userID, exerciseID uuid.UUID) ([]Attempt, error) {
	return s.repo.ListByUserAndExercise(ctx, userID, exerciseID)
}
