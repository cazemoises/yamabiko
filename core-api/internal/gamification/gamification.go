// Package gamification calcula XP, streak diário e badges a partir de tentativas
// (Sec. 5/6 do CLAUDE.md — Fase 6). Substitui o XP hardcoded que existia em
// attempts.xpFor na Fase 4.
package gamification

import (
	"time"

	"github.com/yamabiko/core-api/internal/comparison"
)

type Badge string

const (
	BadgeFirstAttempt Badge = "PRIMEIRA_TENTATIVA"
	BadgeFirstPass    Badge = "PRIMEIRO_PASS"
	BadgeStreak3Days  Badge = "SEQUENCIA_3_DIAS"
	BadgeStreak7Days  Badge = "SEQUENCIA_7_DIAS"
	Badge100XP        Badge = "CEM_XP"
)

// UserStats é o estado persistido de gamificação de um usuário (colunas
// xp_total/current_streak_days/longest_streak_days/last_attempt_date em users).
type UserStats struct {
	XPTotal           int
	CurrentStreakDays int
	LongestStreakDays int
	LastAttemptDate   *time.Time
}

type UpdateResult struct {
	Stats     UserStats
	XPGained  int
	NewBadges []Badge
}

// RecordAttempt atualiza XP/streak a partir de uma nova tentativa e devolve quais
// badges (ainda ausentes em earnedBadges) foram desbloqueadas por ela.
func RecordAttempt(stats UserStats, verdict comparison.Verdict, now time.Time, earnedBadges map[Badge]bool) UpdateResult {
	xpGained := xpByVerdict(verdict)
	isFirstAttemptEver := stats.LastAttemptDate == nil
	newStreak := nextStreak(stats.CurrentStreakDays, stats.LastAttemptDate, now)

	newStats := UserStats{
		XPTotal:           stats.XPTotal + xpGained,
		CurrentStreakDays: newStreak,
		LongestStreakDays: max(stats.LongestStreakDays, newStreak),
		LastAttemptDate:   &now,
	}

	var newBadges []Badge
	award := func(b Badge) {
		if !earnedBadges[b] {
			newBadges = append(newBadges, b)
			earnedBadges[b] = true
		}
	}

	if isFirstAttemptEver {
		award(BadgeFirstAttempt)
	}
	if verdict == comparison.VerdictPass {
		award(BadgeFirstPass)
	}
	if newStats.CurrentStreakDays >= 3 {
		award(BadgeStreak3Days)
	}
	if newStats.CurrentStreakDays >= 7 {
		award(BadgeStreak7Days)
	}
	if newStats.XPTotal >= 100 {
		award(Badge100XP)
	}

	return UpdateResult{Stats: newStats, XPGained: xpGained, NewBadges: newBadges}
}

func xpByVerdict(verdict comparison.Verdict) int {
	switch verdict {
	case comparison.VerdictPass:
		return 10
	case comparison.VerdictPartial:
		return 5
	default:
		return 1
	}
}

func nextStreak(currentStreak int, lastAttemptDate *time.Time, now time.Time) int {
	if lastAttemptDate == nil {
		return 1
	}

	switch daysBetween(toDate(*lastAttemptDate), toDate(now)) {
	case 0:
		return max(currentStreak, 1)
	case 1:
		return currentStreak + 1
	default:
		return 1
	}
}

func toDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func daysBetween(a, b time.Time) int {
	return int(b.Sub(a).Hours() / 24)
}
