package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/auth"
	"github.com/yamabiko/core-api/internal/middleware"
)

func TestRequireAuth_RejectsMissingToken(t *testing.T) {
	issuer := auth.NewJWTIssuer("secret", time.Minute, time.Hour)
	handler := middleware.RequireAuth(issuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, veio %d", rec.Code)
	}
}

func TestRequireAuth_RejectsRefreshTokenAsAccess(t *testing.T) {
	issuer := auth.NewJWTIssuer("secret", time.Minute, time.Hour)
	refreshToken, err := issuer.IssueTokenPair(uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	handler := middleware.RequireAuth(issuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken.RefreshToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, veio %d", rec.Code)
	}
}

func TestRequireAuth_AllowsValidAccessToken(t *testing.T) {
	issuer := auth.NewJWTIssuer("secret", time.Minute, time.Hour)
	userID := uuid.New()
	accessToken, err := issuer.IssueAccessToken(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotUserID uuid.UUID
	handler := middleware.RequireAuth(issuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, _ = middleware.UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, veio %d", rec.Code)
	}
	if gotUserID != userID {
		t.Fatalf("esperava userID %v no contexto, veio %v", userID, gotUserID)
	}
}
