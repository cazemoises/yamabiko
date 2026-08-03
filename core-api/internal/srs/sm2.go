// Package srs implementa spaced repetition via SM-2 pra reintroduzir chunks
// fracos (Sec. 8/9 do CLAUDE.md — "chunk" nesta versão do produto é 1:1 com
// exercises.id, ver migration 0006). TDD obrigatório aqui por decisão do CLAUDE.md.
package srs

import (
	"math"
	"time"
)

type Quality int

const (
	QualityBlackout        Quality = 0
	QualityIncorrect       Quality = 1
	QualityIncorrectEasy   Quality = 2
	QualityCorrectHard     Quality = 3
	QualityCorrectHesitant Quality = 4
	QualityPerfect         Quality = 5
)

type Status string

const (
	StatusNovo      Status = "NOVO"
	StatusEmReforco Status = "EM_REFORCO"
	StatusSolido    Status = "SOLIDO"
)

const (
	defaultEasinessFactor = 2.5
	minEasinessFactor     = 1.3
	solidoRepetitions     = 3
)

// Card é o estado persistido de um chunk pra um usuário (user_chunk_progress).
type Card struct {
	EasinessFactor float64
	IntervalDays   int
	Repetitions    int
}

type Review struct {
	Card         Card
	NextReviewAt time.Time
	Status       Status
}

// Schedule aplica SM-2: recebe o estado atual do card e a qualidade da resposta
// (0-5), devolve o novo estado e a data da próxima revisão.
func Schedule(card Card, quality Quality, now time.Time) Review {
	if card.EasinessFactor == 0 {
		card.EasinessFactor = defaultEasinessFactor
	}

	next := card

	if quality < QualityCorrectHard {
		next.Repetitions = 0
		next.IntervalDays = 1
	} else {
		switch card.Repetitions {
		case 0:
			next.IntervalDays = 1
		case 1:
			next.IntervalDays = 6
		default:
			next.IntervalDays = int(math.Round(float64(card.IntervalDays) * card.EasinessFactor))
		}
		next.Repetitions = card.Repetitions + 1
	}

	q := float64(quality)
	next.EasinessFactor = math.Max(
		minEasinessFactor,
		card.EasinessFactor+(0.1-(5-q)*(0.08+(5-q)*0.02)),
	)

	return Review{
		Card:         next,
		NextReviewAt: now.AddDate(0, 0, next.IntervalDays),
		Status:       statusFor(next),
	}
}

// statusFor só distingue EM_REFORCO/SOLIDO — Schedule só roda depois de uma
// revisão de verdade, então o card por definição já deixou de ser NOVO
// (esse status é o default da migration, atribuído fora do pacote).
func statusFor(card Card) Status {
	if card.Repetitions >= solidoRepetitions {
		return StatusSolido
	}
	return StatusEmReforco
}

// QualityFromScore traduz o similarity_score da engine de comparação (Sec. 3)
// numa qualidade SM-2, usando os mesmos limiares de PASS/PARTIAL/FAIL.
func QualityFromScore(score float64) Quality {
	switch {
	case score >= 0.85:
		return QualityPerfect
	case score >= 0.6:
		return QualityCorrectHard
	default:
		return QualityIncorrect
	}
}
