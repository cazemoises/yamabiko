package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

// RequireAuth vivia em internal/middleware, mas PinSetup (Sec. 4 do pedido
// do usuário: "protegido pelo JWT normal") precisa do userID de dentro do
// próprio pacote auth — e middleware já importa auth, então auth não pode
// importar middleware de volta. Movido pra cá (lógica idêntica, sem
// duplicar) e internal/middleware passa a só reexportar pros chamadores
// existentes (users, tts, dashboard, attempts, httpserver).
func RequireAuth(issuer *JWTIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				writeUnauthorized(w, "token de acesso ausente")
				return
			}

			claims, err := issuer.Parse(token, TokenTypeAccess)
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
