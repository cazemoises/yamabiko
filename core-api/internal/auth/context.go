package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

// HeaderRemoteEmail/HeaderRemoteName são os headers de identidade que o
// Pangolin injeta em requisições que passaram pelo seu SSO antes de
// encaminhar pro core-api (ver
// docs.pangolin.net/manage/access-control/forwarded-headers — só existem
// pra SSO ou shareable link vinculado a usuário; PIN/senha do próprio
// Pangolin não geram identidade nenhuma). Confiados sem verificação
// adicional aqui porque o core-api não deve ser alcançável de nenhuma outra
// forma (nenhuma porta publicada em produção, só a rede interna que o
// Pangolin/newt atravessa — ver docker-compose.prod.yml).
const (
	HeaderRemoteEmail = "Remote-Email"
	HeaderRemoteName  = "Remote-Name"
)

// RequireAuth troca o antigo Bearer JWT por identidade confiada via header
// do Pangolin (Sec. pedida pelo usuário: "não quero mais login com pin ou
// email e senha, quero usar os headers do pangolin"). Sem Remote-Email
// (proxy não autenticou, ou alguém contornou o Pangolin) -> 401. Cada
// pessoa loga no Pangolin com identidade própria (SSO por pessoa,
// confirmado com o usuário) — Remote-Email mapeia 1:1 pro usuário
// correspondente no yamabiko, criado automaticamente no primeiro acesso.
func RequireAuth(repo UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			email := strings.TrimSpace(r.Header.Get(HeaderRemoteEmail))
			if email == "" {
				writeAuthError(w, http.StatusUnauthorized, "identidade do Pangolin ausente (Remote-Email)")
				return
			}
			name := strings.TrimSpace(r.Header.Get(HeaderRemoteName))
			if name == "" {
				name = email
			}

			userID, err := repo.FindOrCreateByEmail(r.Context(), email, name)
			if err != nil {
				writeAuthError(w, http.StatusInternalServerError, "erro ao resolver identidade")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return id, ok
}
