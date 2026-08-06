package auth

import (
	"time"

	"github.com/google/uuid"
)

const (
	pinMaxFailedAttempts = 5
	pinLockDuration      = 15 * time.Minute
)

// PinProfile é a projeção pública de GET /auth/profiles — nunca inclui
// email, pin_hash nem qualquer dado sensível (Sec. 2 do pedido do usuário).
type PinProfile struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	AccentColor *string   `json:"accent_color,omitempty"`
}

// ErrPinInvalid cobre TODOS os motivos de falha que não são lockout: user_id
// desconhecido, usuário sem pin_hash configurado, ou PIN errado — de
// propósito, pra não revelar ao chamador qual dos três foi (Sec. 3.a do
// pedido do usuário). AttemptsRemaining só é preenchido no caso de "PIN
// errado pra uma conta real" (Sec. 3.d permite informar tentativas
// restantes); fica nil nos outros dois casos, que nunca chegam a contar
// tentativa nenhuma.
type ErrPinInvalid struct {
	AttemptsRemaining *int
}

func (e *ErrPinInvalid) Error() string { return "PIN inválido" }

// ErrPinLocked é devolvido tanto quando a tentativa atual acabou de
// disparar o lockout quanto quando já havia um lockout em andamento.
type ErrPinLocked struct {
	RetryAfter time.Duration
}

func (e *ErrPinLocked) Error() string { return "conta temporariamente bloqueada" }

// pinAttemptOutcome é o resultado puro de avaliar uma tentativa de PIN —
// separado de PinLogin (que faz I/O) pra ficar testável sem repositório
// nem tempo de parede, mesmo padrão de srs.Schedule/gamification.RecordAttempt
// (funções puras recebendo `now` como parâmetro).
type pinAttemptOutcome struct {
	success           bool
	locked            bool
	newFailedAttempts int
	newLockedUntil    *time.Time
	attemptsRemaining int
}

func evaluatePinAttempt(currentFailedAttempts int, pinCorrect bool, now time.Time) pinAttemptOutcome {
	if pinCorrect {
		return pinAttemptOutcome{success: true}
	}

	attempts := currentFailedAttempts + 1
	if attempts >= pinMaxFailedAttempts {
		lockedUntil := now.Add(pinLockDuration)
		return pinAttemptOutcome{locked: true, newFailedAttempts: 0, newLockedUntil: &lockedUntil}
	}
	return pinAttemptOutcome{newFailedAttempts: attempts, attemptsRemaining: pinMaxFailedAttempts - attempts}
}
