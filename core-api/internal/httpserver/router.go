package httpserver

import (
	"net"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/yamabiko/core-api/internal/attempts"
	"github.com/yamabiko/core-api/internal/auth"
	"github.com/yamabiko/core-api/internal/dashboard"
	"github.com/yamabiko/core-api/internal/exercises"
	appmiddleware "github.com/yamabiko/core-api/internal/middleware"
	"github.com/yamabiko/core-api/internal/scenarios"
	"github.com/yamabiko/core-api/internal/tts"
	"github.com/yamabiko/core-api/internal/users"
)

func NewRouter(
	authRepo auth.UserRepository,
	exercisesHandler *exercises.Handler,
	attemptsHandler *attempts.Handler,
	usersHandler *users.Handler,
	dashboardHandler *dashboard.Handler,
	ttsHandler *tts.Handler,
	scenariosHandler *scenarios.Handler,
	corsAllowedOrigins []string,
	corsAllowLocalNetwork bool,
	devFakeRemoteEmail string,
	devFakeRemoteName string,
) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(corsOptions(corsAllowedOrigins, corsAllowLocalNetwork)))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Todas as rotas abaixo exigem o header Remote-Email injetado pelo
	// Pangolin (Sec. pedida pelo usuário — substitui o antigo Bearer JWT,
	// ver internal/auth/context.go). Não há mais /auth/register,
	// /auth/login, /auth/pin-login nem /auth/refresh: a identidade já
	// chega pronta em toda requisição que atravessou o Pangolin.
	r.Group(func(r chi.Router) {
		if devFakeRemoteEmail != "" {
			r.Use(devIdentityOverride(devFakeRemoteEmail, devFakeRemoteName))
		}
		r.Use(appmiddleware.RequireAuth(authRepo))

		r.Route("/users/me", func(r chi.Router) {
			r.Get("/", usersHandler.Me)
			r.Get("/progress", usersHandler.Progress)
			r.Patch("/voice-preference", usersHandler.UpdateVoicePreference)
			r.Patch("/appearance", usersHandler.UpdateAppearance)
		})

		r.Route("/exercises", func(r chi.Router) {
			r.Get("/", exercisesHandler.List)
			r.Get("/{id}", exercisesHandler.Get)
			r.Post("/{id}/attempts", attemptsHandler.Submit)
			r.Get("/{id}/attempts", attemptsHandler.History)
			r.Get("/{id}/reference-audio", ttsHandler.ReferenceAudio)
			r.Post("/{id}/answer", exercisesHandler.Answer)
			r.Post("/{id}/text-attempt", exercisesHandler.TextAttempt)
		})

		r.Route("/scenarios", func(r chi.Router) {
			r.Get("/", scenariosHandler.List)
			r.Get("/{id}", scenariosHandler.Detail)
		})

		r.Route("/tts/voices", func(r chi.Router) {
			r.Get("/", ttsHandler.Voices)
			r.Get("/{voice_id}/preview", ttsHandler.VoicePreview)
		})

		r.Get("/dashboard/heatmap", dashboardHandler.Heatmap)
	})

	return r
}

// devIdentityOverride simula o que o Pangolin faz em produção — só usada
// em dev local sem Pangolin na frente (env DEV_FAKE_REMOTE_EMAIL, ver
// internal/config/config.go). Sobrescreve qualquer Remote-Email/Remote-Name
// que porventura já viesse na requisição, então nunca setar essa env var
// atrás de um Pangolin de verdade.
func devIdentityOverride(email, name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set(auth.HeaderRemoteEmail, email)
			if name != "" {
				r.Header.Set(auth.HeaderRemoteName, name)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// corsOptions é a fonte única de verdade da config de CORS — extraída pra
// função própria pra ser testável sem precisar instanciar todos os handlers
// reais do NewRouter (que exigem repositórios Postgres). AllowedMethods
// precisa listar TODO verbo HTTP usado por alguma rota: diferente do que a
// intuição sugere, o go-chi/cors (rs/cors por baixo) confere AllowedMethods
// não só no preflight (OPTIONS), mas também na requisição real — um método
// fora da lista faz a resposta real sair 200 (a rota roda normal) mas SEM
// Access-Control-Allow-Origin, e o browser bloqueia a resposta no cliente
// mesmo a requisição tendo chegado ao servidor. Foi exatamente o que
// aconteceu com PATCH /users/me/voice-preference: a rota existia e
// respondia 200, só faltava "PATCH" nesta lista.
//
// allowLocalNetwork (env CORS_ALLOW_LOCAL_NETWORK, dev-only opt-in — Sec.
// pedida pelo usuário pra testar o `web` a partir do celular na mesma
// Wi-Fi) troca AllowedOrigins por AllowOriginFunc: além da whitelist
// estática, aceita qualquer origin http:// cujo host resolva pra um IP de
// rede privada (RFC1918) ou loopback, sem precisar fixar o IP da máquina
// (que muda a cada rede/DHCP). Quando desligado (default), o
// comportamento é idêntico ao de antes — só AllowedOrigins, nada de
// AllowOriginFunc.
func corsOptions(allowedOrigins []string, allowLocalNetwork bool) cors.Options {
	opts := cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	if !allowLocalNetwork {
		opts.AllowedOrigins = allowedOrigins
		return opts
	}

	static := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		static[origin] = true
	}
	opts.AllowOriginFunc = func(_ *http.Request, origin string) bool {
		return static[origin] || isLocalNetworkOrigin(origin)
	}
	return opts
}

// isLocalNetworkOrigin reconhece origins como http://192.168.1.42:5173 —
// qualquer host que seja um IP literal de rede privada ou loopback. Não
// aceita HTTPS de propósito (dev local só, sem certificado válido pro IP
// da máquina) nem hostnames (só IP literal — evita abrir pra qualquer
// domínio que por acaso resolva pra uma rede privada via DNS rebinding).
func isLocalNetworkOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "http" {
		return false
	}
	ip := net.ParseIP(u.Hostname())
	if ip == nil {
		return false
	}
	return ip.IsPrivate() || ip.IsLoopback()
}
