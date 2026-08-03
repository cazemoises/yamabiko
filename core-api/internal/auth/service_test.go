package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/auth"
)

type fakeRepo struct {
	byEmail map[string]*auth.User
	byID    map[uuid.UUID]*auth.User
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byEmail: map[string]*auth.User{}, byID: map[uuid.UUID]*auth.User{}}
}

func (f *fakeRepo) Create(_ context.Context, user *auth.User) error {
	f.byEmail[user.Email] = user
	f.byID[user.ID] = user
	return nil
}

func (f *fakeRepo) FindByEmail(_ context.Context, email string) (*auth.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return nil, auth.ErrUserNotFound
	}
	return u, nil
}

func (f *fakeRepo) FindByID(_ context.Context, id uuid.UUID) (*auth.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, auth.ErrUserNotFound
	}
	return u, nil
}

func newTestService() (*auth.Service, *fakeRepo) {
	repo := newFakeRepo()
	issuer := auth.NewJWTIssuer("test-secret", time.Minute, time.Hour)
	return auth.NewService(repo, issuer), repo
}

func TestRegister_CreatesUserAndReturnsTokens(t *testing.T) {
	svc, repo := newTestService()

	tokens, err := svc.Register(context.Background(), "moises@example.com", "senha123", "Moisés")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("esperava tokens não vazios")
	}
	if _, ok := repo.byEmail["moises@example.com"]; !ok {
		t.Fatal("usuário não foi persistido")
	}
}

func TestRegister_RejectsDuplicateEmail(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "moises@example.com", "senha123", "Moisés"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.Register(ctx, "moises@example.com", "outrasenha", "Outro Nome")
	if !errors.Is(err, auth.ErrEmailTaken) {
		t.Fatalf("esperava ErrEmailTaken, veio: %v", err)
	}
}

func TestLogin_SucceedsWithCorrectPassword(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	if _, err := svc.Register(ctx, "moises@example.com", "senha123", "Moisés"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tokens, err := svc.Login(ctx, "moises@example.com", "senha123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Fatal("esperava access token")
	}
}

func TestLogin_FailsWithWrongPassword(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	if _, err := svc.Register(ctx, "moises@example.com", "senha123", "Moisés"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.Login(ctx, "moises@example.com", "senhaerrada")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("esperava ErrInvalidCredentials, veio: %v", err)
	}
}

func TestLogin_FailsWithUnknownEmail(t *testing.T) {
	svc, _ := newTestService()
	_, err := svc.Login(context.Background(), "ninguem@example.com", "senha123")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("esperava ErrInvalidCredentials, veio: %v", err)
	}
}

func TestRefresh_IssuesNewAccessTokenForValidRefreshToken(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	tokens, err := svc.Register(ctx, "moises@example.com", "senha123", "Moisés")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	accessToken, err := svc.Refresh(ctx, tokens.RefreshToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accessToken == "" {
		t.Fatal("esperava access token não vazio")
	}
}

func TestRefresh_RejectsAccessTokenUsedAsRefresh(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	tokens, err := svc.Register(ctx, "moises@example.com", "senha123", "Moisés")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.Refresh(ctx, tokens.AccessToken)
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("esperava ErrInvalidToken, veio: %v", err)
	}
}
