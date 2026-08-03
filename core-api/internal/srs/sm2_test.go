package srs_test

import (
	"testing"
	"time"

	"github.com/yamabiko/core-api/internal/srs"
)

var fixedNow = time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)

func TestSchedule_FirstReview_Perfect(t *testing.T) {
	review := srs.Schedule(srs.Card{}, srs.QualityPerfect, fixedNow)

	if review.Card.Repetitions != 1 {
		t.Fatalf("esperava repetitions 1, veio %d", review.Card.Repetitions)
	}
	if review.Card.IntervalDays != 1 {
		t.Fatalf("esperava interval 1, veio %d", review.Card.IntervalDays)
	}
	if review.Card.EasinessFactor != 2.6 {
		t.Fatalf("esperava EF 2.6, veio %v", review.Card.EasinessFactor)
	}
	if review.Status != srs.StatusEmReforco {
		t.Fatalf("esperava EM_REFORCO, veio %v", review.Status)
	}
	if !review.NextReviewAt.Equal(fixedNow.AddDate(0, 0, 1)) {
		t.Fatalf("esperava next_review_at em +1 dia, veio %v", review.NextReviewAt)
	}
}

func TestSchedule_SecondReview_Perfect(t *testing.T) {
	card := srs.Card{EasinessFactor: 2.6, IntervalDays: 1, Repetitions: 1}
	review := srs.Schedule(card, srs.QualityPerfect, fixedNow)

	if review.Card.Repetitions != 2 {
		t.Fatalf("esperava repetitions 2, veio %d", review.Card.Repetitions)
	}
	if review.Card.IntervalDays != 6 {
		t.Fatalf("esperava interval 6, veio %d", review.Card.IntervalDays)
	}
	if review.Card.EasinessFactor != 2.7 {
		t.Fatalf("esperava EF 2.7, veio %v", review.Card.EasinessFactor)
	}
	if review.Status != srs.StatusEmReforco {
		t.Fatalf("esperava EM_REFORCO, veio %v", review.Status)
	}
}

func TestSchedule_ThirdReview_Perfect_GraduatesToSolido(t *testing.T) {
	card := srs.Card{EasinessFactor: 2.7, IntervalDays: 6, Repetitions: 2}
	review := srs.Schedule(card, srs.QualityPerfect, fixedNow)

	if review.Card.Repetitions != 3 {
		t.Fatalf("esperava repetitions 3, veio %d", review.Card.Repetitions)
	}
	if review.Card.IntervalDays != 16 { // round(6 * 2.7) = round(16.2) = 16
		t.Fatalf("esperava interval 16, veio %d", review.Card.IntervalDays)
	}
	if review.Status != srs.StatusSolido {
		t.Fatalf("esperava SOLIDO, veio %v", review.Status)
	}
}

func TestSchedule_FailedReview_ResetsRepetitionsAndInterval(t *testing.T) {
	card := srs.Card{EasinessFactor: 2.8, IntervalDays: 16, Repetitions: 3} // era SOLIDO
	review := srs.Schedule(card, srs.QualityIncorrect, fixedNow)

	if review.Card.Repetitions != 0 {
		t.Fatalf("esperava repetitions 0 após falha, veio %d", review.Card.Repetitions)
	}
	if review.Card.IntervalDays != 1 {
		t.Fatalf("esperava interval 1 após falha, veio %d", review.Card.IntervalDays)
	}
	if review.Status != srs.StatusEmReforco {
		t.Fatalf("esperava voltar pra EM_REFORCO após falha, veio %v", review.Status)
	}
}

func TestSchedule_EasinessFactorNeverGoesBelowMinimum(t *testing.T) {
	card := srs.Card{EasinessFactor: 1.3, IntervalDays: 1, Repetitions: 0}
	review := srs.Schedule(card, srs.QualityBlackout, fixedNow)

	if review.Card.EasinessFactor != 1.3 {
		t.Fatalf("esperava EF clampado em 1.3, veio %v", review.Card.EasinessFactor)
	}
}

func TestQualityFromScore_MatchesComparisonThresholds(t *testing.T) {
	cases := []struct {
		score   float64
		quality srs.Quality
	}{
		{1.0, srs.QualityPerfect},
		{0.85, srs.QualityPerfect},
		{0.84, srs.QualityCorrectHard},
		{0.6, srs.QualityCorrectHard},
		{0.59, srs.QualityIncorrect},
		{0.0, srs.QualityIncorrect},
	}

	for _, c := range cases {
		got := srs.QualityFromScore(c.score)
		if got != c.quality {
			t.Errorf("score %v: esperava quality %v, veio %v", c.score, c.quality, got)
		}
	}
}
