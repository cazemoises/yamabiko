package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidCredentials = errors.New("credenciais inválidas")

type Service struct {
	repo   UserRepository
	tokens *JWTIssuer
}

func NewService(repo UserRepository, tokens *JWTIssuer) *Service {
	return &Service{repo: repo, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, email, password, name string) (*TokenPair, error) {
	_, err := s.repo.FindByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:               uuid.New(),
		Email:            email,
		PasswordHash:     hash,
		Name:             name,
		CurrentSprintDay: 1,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.tokens.IssueTokenPair(user.ID)
}

func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !CheckPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	return s.tokens.IssueTokenPair(user.ID)
}

// Profiles alimenta GET /auth/profiles — só perfis com PIN configurado.
func (s *Service) Profiles(ctx context.Context) ([]PinProfile, error) {
	return s.repo.ListPinEnabledProfiles(ctx)
}

// PinLogin implementa o fluxo da Sec. 3 do pedido do usuário. `now` é
// parâmetro explícito (não time.Now() interno) pra deixar
// evaluatePinAttempt/o fluxo de lockout testável sem tempo de parede —
// mesmo padrão de srs.Schedule.
func (s *Service) PinLogin(ctx context.Context, userID uuid.UUID, pin string, now time.Time) (*TokenPair, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user.PinHash == nil {
		return nil, &ErrPinInvalid{}
	}

	if user.PinLockedUntil != nil && user.PinLockedUntil.After(now) {
		return nil, &ErrPinLocked{RetryAfter: user.PinLockedUntil.Sub(now)}
	}

	correct := CheckPassword(*user.PinHash, pin)
	outcome := evaluatePinAttempt(user.PinFailedAttempts, correct, now)

	if outcome.success {
		if err := s.repo.UpdatePinAuthState(ctx, userID, 0, nil); err != nil {
			return nil, err
		}
		return s.tokens.IssueTokenPair(userID)
	}

	if err := s.repo.UpdatePinAuthState(ctx, userID, outcome.newFailedAttempts, outcome.newLockedUntil); err != nil {
		return nil, err
	}
	if outcome.locked {
		return nil, &ErrPinLocked{RetryAfter: pinLockDuration}
	}
	remaining := outcome.attemptsRemaining
	return nil, &ErrPinInvalid{AttemptsRemaining: &remaining}
}

// SetPin serve POST /auth/pin-setup — só alcançável autenticado por senha
// (Sec. 5 do pedido do usuário: senha continua sendo a porta única de
// configuração/reset de PIN). Validação de formato (6 dígitos) é
// responsabilidade do Handler, não do Service.
func (s *Service) SetPin(ctx context.Context, userID uuid.UUID, pin string) error {
	hash, err := HashPassword(pin)
	if err != nil {
		return err
	}
	return s.repo.SetPinHash(ctx, userID, hash)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.tokens.Parse(refreshToken, TokenTypeRefresh)
	if err != nil {
		return "", ErrInvalidToken
	}
	if _, err := s.repo.FindByID(ctx, claims.UserID); err != nil {
		return "", ErrInvalidToken
	}
	return s.tokens.IssueAccessToken(claims.UserID)
}
