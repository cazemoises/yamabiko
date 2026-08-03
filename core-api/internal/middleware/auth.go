package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/auth"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

func RequireAuth(issuer *auth.JWTIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				writeUnauthorized(w, "token de acesso ausente")
				return
			}

			claims, err := issuer.Parse(token, auth.TokenTypeAccess)
			if err != nil {
				writeUnauthorized(w, "token de acesso inválido")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return id, ok
}
