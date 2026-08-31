package auth_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/auth"
)

// fakeUserRepository só registra as chamadas recebidas — não precisa de
// Postgres pra testar o middleware isoladamente.
type fakeUserRepository struct {
	byEmail  map[string]uuid.UUID
	lastName string
	err      error
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{byEmail: map[string]uuid.UUID{}}
}

func (f *fakeUserRepository) FindOrCreateByEmail(_ context.Context, email, name string) (uuid.UUID, error) {
	if f.err != nil {
		return uuid.Nil, f.err
	}
	f.lastName = name
	if id, ok := f.byEmail[email]; ok {
		return id, nil
	}
	id := uuid.New()
	f.byEmail[email] = id
	return id, nil
}

func TestRequireAuth_RejectsMissingRemoteEmail(t *testing.T) {
	repo := newFakeUserRepository()
	handler := auth.RequireAuth(repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, veio %d", rec.Code)
	}
}

func TestRequireAuth_RejectsBlankRemoteEmail(t *testing.T) {
	repo := newFakeUserRepository()
	handler := auth.RequireAuth(repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set(auth.HeaderRemoteEmail, "   ")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, veio %d", rec.Code)
	}
}

func TestRequireAuth_ResolvesUserFromRemoteEmail(t *testing.T) {
	repo := newFakeUserRepository()
	wantID := uuid.New()
	repo.byEmail["vitoria@example.com"] = wantID

	var gotUserID uuid.UUID
	handler := auth.RequireAuth(repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, _ = auth.UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set(auth.HeaderRemoteEmail, "vitoria@example.com")
	req.Header.Set(auth.HeaderRemoteName, "Vitória")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, veio %d", rec.Code)
	}
	if gotUserID != wantID {
		t.Fatalf("esperava userID %v no contexto, veio %v", wantID, gotUserID)
	}
}

func TestRequireAuth_CreatesUserOnFirstAccess(t *testing.T) {
	repo := newFakeUserRepository()
	handler := auth.RequireAuth(repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set(auth.HeaderRemoteEmail, "novo@example.com")
	req.Header.Set(auth.HeaderRemoteName, "Novo Usuário")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, veio %d", rec.Code)
	}
	if _, ok := repo.byEmail["novo@example.com"]; !ok {
		t.Fatal("esperava usuário criado pro email novo")
	}
	if repo.lastName != "Novo Usuário" {
		t.Fatalf("esperava name %q repassado pro repo, veio %q", "Novo Usuário", repo.lastName)
	}
}

func TestRequireAuth_FallsBackNameToEmailWhenRemoteNameMissing(t *testing.T) {
	repo := newFakeUserRepository()
	handler := auth.RequireAuth(repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set(auth.HeaderRemoteEmail, "semnome@example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if repo.lastName != "semnome@example.com" {
		t.Fatalf("esperava name igual ao email como fallback, veio %q", repo.lastName)
	}
}

func TestRequireAuth_RepositoryErrorIsInternalError(t *testing.T) {
	repo := newFakeUserRepository()
	repo.err = errors.New("db indisponível")
	handler := auth.RequireAuth(repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set(auth.HeaderRemoteEmail, "vitoria@example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("esperava 500, veio %d", rec.Code)
	}
}
