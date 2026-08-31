package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/auth"
	"github.com/yamabiko/core-api/internal/middleware"
)

// fakeUserRepository cobre só o suficiente pra exercitar o reexport de
// middleware.RequireAuth/UserIDFromContext — o comportamento em si já é
// coberto a fundo em internal/auth/context_test.go.
type fakeUserRepository struct{}

func (fakeUserRepository) FindOrCreateByEmail(_ context.Context, _, _ string) (uuid.UUID, error) {
	return uuid.New(), nil
}

func TestRequireAuth_RejectsMissingRemoteEmail(t *testing.T) {
	handler := middleware.RequireAuth(fakeUserRepository{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, veio %d", rec.Code)
	}
}

func TestRequireAuth_AllowsValidRemoteEmail(t *testing.T) {
	var gotUserID uuid.UUID
	handler := middleware.RequireAuth(fakeUserRepository{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, _ = middleware.UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set(auth.HeaderRemoteEmail, "vitoria@example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, veio %d", rec.Code)
	}
	if gotUserID == uuid.Nil {
		t.Fatal("esperava userID não-nulo no contexto")
	}
}
