package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/yamabiko/core-api/internal/attempts"
	"github.com/yamabiko/core-api/internal/auth"
	"github.com/yamabiko/core-api/internal/exercises"
	appmiddleware "github.com/yamabiko/core-api/internal/middleware"
)

func NewRouter(
	authHandler *auth.Handler,
	tokens *auth.JWTIssuer,
	exercisesHandler *exercises.Handler,
	attemptsHandler *attempts.Handler,
) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
	})

	// Todas as rotas abaixo exigem Authorization: Bearer {access_token} (Sec. 4 do CLAUDE.md).
	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.RequireAuth(tokens))

		r.Get("/users/me", meHandler)

		r.Route("/exercises", func(r chi.Router) {
			r.Get("/", exercisesHandler.List)
			r.Get("/{id}", exercisesHandler.Get)
			r.Post("/{id}/attempts", attemptsHandler.Submit)
			r.Get("/{id}/attempts", attemptsHandler.History)
		})
	})

	return r
}

// meHandler prova que o middleware de auth funciona ponta a ponta.
// Perfil completo (XP, streak, sprint) chega na Fase 3+ junto do domínio de users.
func meHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"id":"` + userID.String() + `"}`))
}
