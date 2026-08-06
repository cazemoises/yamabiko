package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/auth"
)

var pinNow = time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)

func parseUUID(t *testing.T, s string) (uuid.UUID, error) {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("unexpected error parsing uuid %q: %v", s, err)
	}
	return id, err
}

func registerAndSetPin(t *testing.T, svc *auth.Service, email, pin string) (userID string) {
	t.Helper()
	ctx := context.Background()
	tokens, err := svc.Register(ctx, email, "senha-de-recuperacao", "Perfil de Teste")
	if err != nil {
		t.Fatalf("unexpected error registering: %v", err)
	}
	claims, err := auth.NewJWTIssuer("test-secret", time.Minute, time.Hour).Parse(tokens.AccessToken, auth.TokenTypeAccess)
	if err != nil {
		t.Fatalf("unexpected error parsing token: %v", err)
	}
	if err := svc.SetPin(ctx, claims.UserID, pin); err != nil {
		t.Fatalf("unexpected error setting pin: %v", err)
	}
	return claims.UserID.String()
}

// newTestServiceSharedIssuer garante que o JWT emitido no registro e o
// usado pra decodificar o user_id (helper acima) compartilham o mesmo
// segredo — newTestService() já usa "test-secret", só documentando aqui.
func TestProfiles_OnlyListsUsersWithPinConfigured(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()

	tokens, err := svc.Register(ctx, "semdpin@example.com", "senha123", "Sem PIN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = tokens

	userIDStr := registerAndSetPin(t, svc, "compin@example.com", "123456")

	profiles, err := svc.Profiles(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("esperava 1 perfil, veio %d", len(profiles))
	}
	if profiles[0].ID.String() != userIDStr {
		t.Fatalf("esperava perfil %s, veio %s", userIDStr, profiles[0].ID)
	}
	if profiles[0].DisplayName != "Perfil de Teste" {
		t.Fatalf("esperava display_name preenchido, veio %q", profiles[0].DisplayName)
	}

	// repo interno não deve ter sido mutado com email/pin_hash na projeção —
	// checagem estrutural: PinProfile não tem esses campos, então o próprio
	// compilador já garante isso; aqui só confirmamos que o usuário sem PIN
	// realmente ficou de fora.
	if _, ok := repo.byEmail["semdpin@example.com"]; !ok {
		t.Fatal("usuário sem PIN deveria continuar existindo no repo, só não listado")
	}
}

func TestPinLogin_SucceedsWithCorrectPin(t *testing.T) {
	svc, _ := newTestService()
	userIDStr := registerAndSetPin(t, svc, "cazé@example.com", "123456")
	userID, _ := parseUUID(t, userIDStr)

	tokens, err := svc.PinLogin(context.Background(), userID, "123456", pinNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("esperava tokens não vazios")
	}
}

func TestPinLogin_WrongPinReturnsGenericErrorWithAttemptsRemaining(t *testing.T) {
	svc, _ := newTestService()
	userIDStr := registerAndSetPin(t, svc, "cazé@example.com", "123456")
	userID, _ := parseUUID(t, userIDStr)

	_, err := svc.PinLogin(context.Background(), userID, "000000", pinNow)
	var invalidErr *auth.ErrPinInvalid
	if !errors.As(err, &invalidErr) {
		t.Fatalf("esperava *auth.ErrPinInvalid, veio: %v", err)
	}
	if invalidErr.AttemptsRemaining == nil || *invalidErr.AttemptsRemaining != 4 {
		t.Fatalf("esperava 4 tentativas restantes, veio: %v", invalidErr.AttemptsRemaining)
	}
}

func TestPinLogin_LocksAfterFiveFailedAttempts(t *testing.T) {
	svc, _ := newTestService()
	userIDStr := registerAndSetPin(t, svc, "cazé@example.com", "123456")
	userID, _ := parseUUID(t, userIDStr)
	ctx := context.Background()

	for i := range 4 {
		_, err := svc.PinLogin(ctx, userID, "000000", pinNow)
		var invalidErr *auth.ErrPinInvalid
		if !errors.As(err, &invalidErr) {
			t.Fatalf("tentativa %d: esperava *auth.ErrPinInvalid, veio: %v", i+1, err)
		}
	}

	_, err := svc.PinLogin(ctx, userID, "000000", pinNow)
	var lockedErr *auth.ErrPinLocked
	if !errors.As(err, &lockedErr) {
		t.Fatalf("esperava *auth.ErrPinLocked na 5ª tentativa errada, veio: %v", err)
	}
	if lockedErr.RetryAfter != 15*time.Minute {
		t.Fatalf("esperava lockout de 15min, veio: %v", lockedErr.RetryAfter)
	}
}

func TestPinLogin_RejectsDuringLockoutEvenWithCorrectPin(t *testing.T) {
	svc, _ := newTestService()
	userIDStr := registerAndSetPin(t, svc, "cazé@example.com", "123456")
	userID, _ := parseUUID(t, userIDStr)
	ctx := context.Background()

	for range 5 {
		_, _ = svc.PinLogin(ctx, userID, "000000", pinNow)
	}

	fiveMinutesLater := pinNow.Add(5 * time.Minute)
	_, err := svc.PinLogin(ctx, userID, "123456", fiveMinutesLater)
	var lockedErr *auth.ErrPinLocked
	if !errors.As(err, &lockedErr) {
		t.Fatalf("esperava *auth.ErrPinLocked mesmo com PIN correto durante lockout, veio: %v", err)
	}
	if lockedErr.RetryAfter != 10*time.Minute {
		t.Fatalf("esperava 10min restantes de lockout, veio: %v", lockedErr.RetryAfter)
	}

	afterLockout := pinNow.Add(15*time.Minute + time.Second)
	tokens, err := svc.PinLogin(ctx, userID, "123456", afterLockout)
	if err != nil {
		t.Fatalf("esperava sucesso depois do lockout expirar, veio: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Fatal("esperava access token")
	}
}

func TestPinLogin_UserWithoutPinConfiguredReturnsGenericError(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	tokens, err := svc.Register(ctx, "semdpin@example.com", "senha123", "Sem PIN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	claims, err := auth.NewJWTIssuer("test-secret", time.Minute, time.Hour).Parse(tokens.AccessToken, auth.TokenTypeAccess)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.PinLogin(ctx, claims.UserID, "123456", pinNow)
	var invalidErr *auth.ErrPinInvalid
	if !errors.As(err, &invalidErr) {
		t.Fatalf("esperava *auth.ErrPinInvalid, veio: %v", err)
	}
	if invalidErr.AttemptsRemaining != nil {
		t.Fatalf("usuário sem PIN configurado não deve contar tentativas, veio: %v", *invalidErr.AttemptsRemaining)
	}
}

func TestPinLogin_UnknownUserIDReturnsGenericError(t *testing.T) {
	svc, _ := newTestService()

	randomID, _ := parseUUID(t, "00000000-0000-0000-0000-000000000000")
	_, err := svc.PinLogin(context.Background(), randomID, "123456", pinNow)
	var invalidErr *auth.ErrPinInvalid
	if !errors.As(err, &invalidErr) {
		t.Fatalf("esperava *auth.ErrPinInvalid, veio: %v", err)
	}
	if invalidErr.AttemptsRemaining != nil {
		t.Fatalf("user_id desconhecido não deve contar tentativas, veio: %v", *invalidErr.AttemptsRemaining)
	}
}

func TestPinLogin_SuccessResetsFailedAttempts(t *testing.T) {
	svc, repo := newTestService()
	userIDStr := registerAndSetPin(t, svc, "cazé@example.com", "123456")
	userID, _ := parseUUID(t, userIDStr)
	ctx := context.Background()

	_, _ = svc.PinLogin(ctx, userID, "000000", pinNow)
	_, _ = svc.PinLogin(ctx, userID, "000000", pinNow)

	if _, err := svc.PinLogin(ctx, userID, "123456", pinNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u := repo.byID[userID]
	if u.PinFailedAttempts != 0 {
		t.Fatalf("esperava pin_failed_attempts=0 após sucesso, veio %d", u.PinFailedAttempts)
	}
	if u.PinLockedUntil != nil {
		t.Fatalf("esperava pin_locked_until=nil após sucesso, veio %v", u.PinLockedUntil)
	}
}
