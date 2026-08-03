package gamification_test

import (
	"slices"
	"testing"
	"time"

	"github.com/yamabiko/core-api/internal/comparison"
	"github.com/yamabiko/core-api/internal/gamification"
)

var today = time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
var yesterday = today.AddDate(0, 0, -1)
var threeDaysAgo = today.AddDate(0, 0, -3)

func TestRecordAttempt_VeryFirstAttempt_AwardsFirstAttemptAndFirstPass(t *testing.T) {
	result := gamification.RecordAttempt(gamification.UserStats{}, comparison.VerdictPass, today, map[gamification.Badge]bool{})

	if result.XPGained != 10 {
		t.Fatalf("esperava XP 10 pra PASS, veio %d", result.XPGained)
	}
	if result.Stats.XPTotal != 10 {
		t.Fatalf("esperava XPTotal 10, veio %d", result.Stats.XPTotal)
	}
	if result.Stats.CurrentStreakDays != 1 || result.Stats.LongestStreakDays != 1 {
		t.Fatalf("esperava streak 1/1, veio %d/%d", result.Stats.CurrentStreakDays, result.Stats.LongestStreakDays)
	}
	if !containsBadge(result.NewBadges, gamification.BadgeFirstAttempt) {
		t.Fatal("esperava BadgeFirstAttempt")
	}
	if !containsBadge(result.NewBadges, gamification.BadgeFirstPass) {
		t.Fatal("esperava BadgeFirstPass")
	}
}

func TestRecordAttempt_NextConsecutiveDay_IncrementsStreak(t *testing.T) {
	stats := gamification.UserStats{XPTotal: 10, CurrentStreakDays: 1, LongestStreakDays: 1, LastAttemptDate: &yesterday}
	earned := map[gamification.Badge]bool{gamification.BadgeFirstAttempt: true}

	result := gamification.RecordAttempt(stats, comparison.VerdictPartial, today, earned)

	if result.XPGained != 5 {
		t.Fatalf("esperava XP 5 pra PARTIAL, veio %d", result.XPGained)
	}
	if result.Stats.CurrentStreakDays != 2 {
		t.Fatalf("esperava streak 2, veio %d", result.Stats.CurrentStreakDays)
	}
	if result.Stats.LongestStreakDays != 2 {
		t.Fatalf("esperava longest streak 2, veio %d", result.Stats.LongestStreakDays)
	}
}

func TestRecordAttempt_SameDay_DoesNotChangeStreak(t *testing.T) {
	earlierToday := today.Add(-2 * time.Hour)
	stats := gamification.UserStats{CurrentStreakDays: 3, LongestStreakDays: 3, LastAttemptDate: &earlierToday}

	result := gamification.RecordAttempt(stats, comparison.VerdictFail, today, map[gamification.Badge]bool{})

	if result.Stats.CurrentStreakDays != 3 {
		t.Fatalf("esperava streak inalterado em 3, veio %d", result.Stats.CurrentStreakDays)
	}
}

func TestRecordAttempt_GapInAttempts_ResetsStreakToOne(t *testing.T) {
	stats := gamification.UserStats{CurrentStreakDays: 5, LongestStreakDays: 5, LastAttemptDate: &threeDaysAgo}

	result := gamification.RecordAttempt(stats, comparison.VerdictFail, today, map[gamification.Badge]bool{})

	if result.Stats.CurrentStreakDays != 1 {
		t.Fatalf("esperava streak resetado pra 1, veio %d", result.Stats.CurrentStreakDays)
	}
	if result.Stats.LongestStreakDays != 5 {
		t.Fatalf("longest streak não deveria cair, veio %d", result.Stats.LongestStreakDays)
	}
}

func TestRecordAttempt_StreakReaches3_AwardsBadge(t *testing.T) {
	stats := gamification.UserStats{CurrentStreakDays: 2, LastAttemptDate: &yesterday}

	result := gamification.RecordAttempt(stats, comparison.VerdictFail, today, map[gamification.Badge]bool{})

	if !containsBadge(result.NewBadges, gamification.BadgeStreak3Days) {
		t.Fatal("esperava BadgeStreak3Days ao atingir streak 3")
	}
}

func TestRecordAttempt_StreakReaches7_AwardsBadge(t *testing.T) {
	stats := gamification.UserStats{CurrentStreakDays: 6, LastAttemptDate: &yesterday}

	result := gamification.RecordAttempt(stats, comparison.VerdictFail, today, map[gamification.Badge]bool{})

	if !containsBadge(result.NewBadges, gamification.BadgeStreak7Days) {
		t.Fatal("esperava BadgeStreak7Days ao atingir streak 7")
	}
}

func TestRecordAttempt_XPCrossing100_AwardsBadge(t *testing.T) {
	stats := gamification.UserStats{XPTotal: 95, LastAttemptDate: &yesterday}

	result := gamification.RecordAttempt(stats, comparison.VerdictPass, today, map[gamification.Badge]bool{})

	if result.Stats.XPTotal != 105 {
		t.Fatalf("esperava XPTotal 105, veio %d", result.Stats.XPTotal)
	}
	if !containsBadge(result.NewBadges, gamification.Badge100XP) {
		t.Fatal("esperava Badge100XP ao cruzar 100 XP")
	}
}

func TestRecordAttempt_AlreadyEarnedBadge_NotAwardedAgain(t *testing.T) {
	stats := gamification.UserStats{XPTotal: 50, LastAttemptDate: &yesterday}
	earned := map[gamification.Badge]bool{gamification.BadgeFirstPass: true}

	result := gamification.RecordAttempt(stats, comparison.VerdictPass, today, earned)

	if containsBadge(result.NewBadges, gamification.BadgeFirstPass) {
		t.Fatal("não deveria reconquistar BadgeFirstPass")
	}
}

func containsBadge(badges []gamification.Badge, target gamification.Badge) bool {
	return slices.Contains(badges, target)
}
