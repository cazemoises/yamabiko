package attempts_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/attempts"
	"github.com/yamabiko/core-api/internal/comparison"
	"github.com/yamabiko/core-api/internal/exercises"
	"github.com/yamabiko/core-api/internal/gamification"
	"github.com/yamabiko/core-api/internal/srs"
	"github.com/yamabiko/core-api/internal/sttclient"
)

type fakeRepo struct {
	created []attempts.Attempt
}

func (f *fakeRepo) Create(_ context.Context, attempt *attempts.Attempt) error {
	f.created = append(f.created, *attempt)
	return nil
}

func (f *fakeRepo) ListByUserAndExercise(_ context.Context, userID, exerciseID uuid.UUID) ([]attempts.Attempt, error) {
	var result []attempts.Attempt
	for _, a := range f.created {
		if a.UserID == userID && a.ExerciseID == exerciseID {
			result = append(result, a)
		}
	}
	return result, nil
}

type fakeTranscriber struct {
	transcript string
	err        error
}

func (f *fakeTranscriber) Transcribe(_ context.Context, _ string, _ io.Reader) (*sttclient.TranscriptionResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &sttclient.TranscriptionResult{Transcript: f.transcript, Language: "ja", Confidence: 0.95}, nil
}

type fakeExerciseFinder struct {
	exercise *exercises.Exercise
	err      error
}

func (f *fakeExerciseFinder) FindByID(_ context.Context, _ uuid.UUID) (*exercises.Exercise, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.exercise, nil
}

type fakeSRSRepo struct {
	saved map[string]srs.Review
}

func newFakeSRSRepo() *fakeSRSRepo {
	return &fakeSRSRepo{saved: map[string]srs.Review{}}
}

func (f *fakeSRSRepo) Get(_ context.Context, userID, chunkID uuid.UUID) (srs.Card, srs.Status, error) {
	if review, ok := f.saved[userID.String()+chunkID.String()]; ok {
		return review.Card, review.Status, nil
	}
	return srs.Card{}, srs.StatusNovo, nil
}

func (f *fakeSRSRepo) Save(_ context.Context, userID, chunkID uuid.UUID, review srs.Review) error {
	f.saved[userID.String()+chunkID.String()] = review
	return nil
}

type fakeUserStatsRepo struct {
	stats  map[uuid.UUID]gamification.UserStats
	badges map[uuid.UUID]map[gamification.Badge]bool
}

func newFakeUserStatsRepo() *fakeUserStatsRepo {
	return &fakeUserStatsRepo{
		stats:  map[uuid.UUID]gamification.UserStats{},
		badges: map[uuid.UUID]map[gamification.Badge]bool{},
	}
}

func (f *fakeUserStatsRepo) FindStatsByID(_ context.Context, userID uuid.UUID) (*gamification.UserStats, error) {
	stats := f.stats[userID]
	return &stats, nil
}

func (f *fakeUserStatsRepo) UpdateStats(_ context.Context, userID uuid.UUID, stats gamification.UserStats) error {
	f.stats[userID] = stats
	return nil
}

func (f *fakeUserStatsRepo) EarnedBadges(_ context.Context, userID uuid.UUID) (map[gamification.Badge]bool, error) {
	badges, ok := f.badges[userID]
	if !ok {
		badges = map[gamification.Badge]bool{}
		f.badges[userID] = badges
	}
	return badges, nil
}

func (f *fakeUserStatsRepo) AwardBadges(_ context.Context, userID uuid.UUID, newBadges []gamification.Badge) error {
	badges := f.badges[userID]
	for _, b := range newBadges {
		badges[b] = true
	}
	return nil
}

type fakePhoneticsRepo struct {
	incremented map[comparison.ErrorPattern]int
}

func newFakePhoneticsRepo() *fakePhoneticsRepo {
	return &fakePhoneticsRepo{incremented: map[comparison.ErrorPattern]int{}}
}

func (f *fakePhoneticsRepo) IncrementPatterns(_ context.Context, _ uuid.UUID, _ string, counts map[comparison.ErrorPattern]int) error {
	for pattern, count := range counts {
		f.incremented[pattern] += count
	}
	return nil
}

func newTestService(repo *fakeRepo, transcriber *fakeTranscriber, finder *fakeExerciseFinder) *attempts.Service {
	return attempts.NewService(repo, transcriber, finder, newFakeSRSRepo(), newFakeUserStatsRepo(), newFakePhoneticsRepo())
}

func TestSubmit_PersistsAttemptAndReturnsXPForExactMatch(t *testing.T) {
	repo := &fakeRepo{}
	transcriber := &fakeTranscriber{transcript: "おはよう"}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{ExpectedTranscript: "おはよう"}}
	service := newTestService(repo, transcriber, finder)

	userID, exerciseID := uuid.New(), uuid.New()
	result, err := service.Submit(context.Background(), userID, exerciseID, "attempt.wav", strings.NewReader("audio"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Verdict != comparison.VerdictPass {
		t.Fatalf("esperava PASS, veio %v", result.Verdict)
	}
	if result.XPGained != 10 {
		t.Fatalf("esperava 10 XP pra PASS, veio %d", result.XPGained)
	}
	if len(repo.created) != 1 {
		t.Fatalf("esperava 1 attempt persistido, veio %d", len(repo.created))
	}
	if repo.created[0].UserID != userID || repo.created[0].ExerciseID != exerciseID {
		t.Fatal("attempt persistido com userID/exerciseID errados")
	}
}

func TestSubmit_ReturnsLowerXPForFail(t *testing.T) {
	repo := &fakeRepo{}
	transcriber := &fakeTranscriber{transcript: "ぜんぜんちがう"}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{ExpectedTranscript: "おはよう"}}
	service := newTestService(repo, transcriber, finder)

	result, err := service.Submit(context.Background(), uuid.New(), uuid.New(), "attempt.wav", strings.NewReader("audio"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != comparison.VerdictFail {
		t.Fatalf("esperava FAIL, veio %v", result.Verdict)
	}
	if result.XPGained != 1 {
		t.Fatalf("esperava 1 XP pra FAIL, veio %d", result.XPGained)
	}
}

func TestSubmit_FirstAttemptEver_ReturnsFirstAttemptBadge(t *testing.T) {
	repo := &fakeRepo{}
	transcriber := &fakeTranscriber{transcript: "おはよう"}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{ExpectedTranscript: "おはよう"}}
	service := newTestService(repo, transcriber, finder)

	result, err := service.Submit(context.Background(), uuid.New(), uuid.New(), "attempt.wav", strings.NewReader("audio"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, b := range result.NewBadges {
		if b == gamification.BadgeFirstAttempt {
			found = true
		}
	}
	if !found {
		t.Fatalf("esperava BadgeFirstAttempt, veio %v", result.NewBadges)
	}
}

func TestSubmit_PropagatesExerciseNotFound(t *testing.T) {
	repo := &fakeRepo{}
	transcriber := &fakeTranscriber{transcript: "おはよう"}
	finder := &fakeExerciseFinder{err: exercises.ErrExerciseNotFound}
	service := newTestService(repo, transcriber, finder)

	_, err := service.Submit(context.Background(), uuid.New(), uuid.New(), "attempt.wav", strings.NewReader("audio"))
	if !errors.Is(err, exercises.ErrExerciseNotFound) {
		t.Fatalf("esperava ErrExerciseNotFound, veio %v", err)
	}
	if len(repo.created) != 0 {
		t.Fatal("não deveria persistir attempt quando exercício não existe")
	}
}

func TestSubmit_DoesNotPersistWhenTranscriptionFails(t *testing.T) {
	repo := &fakeRepo{}
	transcriber := &fakeTranscriber{err: errors.New("stt-service indisponível")}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{ExpectedTranscript: "おはよう"}}
	service := newTestService(repo, transcriber, finder)

	_, err := service.Submit(context.Background(), uuid.New(), uuid.New(), "attempt.wav", strings.NewReader("audio"))
	if err == nil {
		t.Fatal("esperava erro quando stt-service falha")
	}
	if len(repo.created) != 0 {
		t.Fatal("não deveria persistir attempt quando transcrição falha")
	}
}

func TestHistory_ReturnsOnlyAttemptsForUserAndExercise(t *testing.T) {
	repo := &fakeRepo{}
	transcriber := &fakeTranscriber{transcript: "おはよう"}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{ExpectedTranscript: "おはよう"}}
	service := newTestService(repo, transcriber, finder)

	userID, exerciseID := uuid.New(), uuid.New()
	otherExerciseID := uuid.New()

	if _, err := service.Submit(context.Background(), userID, exerciseID, "a.wav", strings.NewReader("audio")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := service.Submit(context.Background(), userID, otherExerciseID, "b.wav", strings.NewReader("audio")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	history, err := service.History(context.Background(), userID, exerciseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("esperava 1 attempt no histórico, veio %d", len(history))
	}
}
