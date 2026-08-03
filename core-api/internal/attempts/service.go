package attempts

import (
	"context"
	"io"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/comparison"
	"github.com/yamabiko/core-api/internal/exercises"
	"github.com/yamabiko/core-api/internal/sttclient"
)

type Transcriber interface {
	Transcribe(ctx context.Context, filename string, audio io.Reader) (*sttclient.TranscriptionResult, error)
}

type ExerciseFinder interface {
	FindByID(ctx context.Context, id uuid.UUID) (*exercises.Exercise, error)
}

type Service struct {
	repo           Repository
	transcriber    Transcriber
	exerciseFinder ExerciseFinder
}

func NewService(repo Repository, transcriber Transcriber, exerciseFinder ExerciseFinder) *Service {
	return &Service{repo: repo, transcriber: transcriber, exerciseFinder: exerciseFinder}
}

type SubmitResult struct {
	Transcript string
	Score      float64
	Verdict    comparison.Verdict
	Diff       []comparison.DiffEntry
	XPGained   int
}

func (s *Service) Submit(ctx context.Context, userID, exerciseID uuid.UUID, filename string, audio io.Reader) (*SubmitResult, error) {
	exercise, err := s.exerciseFinder.FindByID(ctx, exerciseID)
	if err != nil {
		return nil, err
	}

	transcription, err := s.transcriber.Transcribe(ctx, filename, audio)
	if err != nil {
		return nil, err
	}

	result := comparison.Compare(exercise.ExpectedTranscript, transcription.Transcript)

	attempt := &Attempt{
		ID:              uuid.New(),
		UserID:          userID,
		ExerciseID:      exerciseID,
		AudioTranscript: transcription.Transcript,
		SimilarityScore: result.SimilarityScore,
		Verdict:         result.Verdict,
		PhoneticDiff:    result.PhoneticDiff,
	}
	if err := s.repo.Create(ctx, attempt); err != nil {
		return nil, err
	}

	return &SubmitResult{
		Transcript: transcription.Transcript,
		Score:      result.SimilarityScore,
		Verdict:    result.Verdict,
		Diff:       result.PhoneticDiff,
		XPGained:   xpFor(result.Verdict),
	}, nil
}

func (s *Service) History(ctx context.Context, userID, exerciseID uuid.UUID) ([]Attempt, error) {
	return s.repo.ListByUserAndExercise(ctx, userID, exerciseID)
}

// xpFor é um cálculo mínimo de XP por veredito, só pra fechar o contrato de
// resposta da Sec. 4 (`xp_gained`). Gamificação completa — streak, multiplicadores,
// badges — é escopo da Fase 6.
func xpFor(verdict comparison.Verdict) int {
	switch verdict {
	case comparison.VerdictPass:
		return 10
	case comparison.VerdictPartial:
		return 5
	default:
		return 1
	}
}
