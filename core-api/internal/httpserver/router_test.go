package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

// newTestCORSRouter monta um router mínimo com a MESMA config de CORS
// (corsOptions) usada pelo NewRouter de produção, sem precisar instanciar
// nenhum handler real — o objetivo é testar a config de CORS em si, não o
// roteamento de negócio (isso já é coberto pelos testes de cada pacote de
// handler).
func newTestCORSRouter(allowLocalNetwork bool) http.Handler {
	r := chi.NewRouter()
	r.Use(cors.Handler(corsOptions([]string{"http://localhost:5173"}, allowLocalNetwork)))
	r.Get("/probe", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Post("/probe", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Put("/probe", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Patch("/probe", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Delete("/probe", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	return r
}

// TestCORS_ActualRequest_SetsAllowOriginForEveryAllowedMethod é o teste de
// regressão pedido: cobre cada verbo em AllowedMethods com uma requisição
// REAL (não preflight) com Origin setado, e falha se
// Access-Control-Allow-Origin não vier na resposta — é exatamente essa
// ausência de header (não um 403/405) que bloqueia a chamada no browser.
// Se um verbo novo for usado por uma rota futura sem entrar em
// corsOptions.AllowedMethods, adicionar o método aqui pega a regressão.
func TestCORS_ActualRequest_SetsAllowOriginForEveryAllowedMethod(t *testing.T) {
	router := newTestCORSRouter(false)

	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/probe", nil)
			req.Header.Set("Origin", "http://localhost:5173")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
				t.Fatalf("%s: esperava Access-Control-Allow-Origin='http://localhost:5173', veio %q (status %d)", method, got, rec.Code)
			}
		})
	}
}

// TestCORS_Preflight_PatchIsInAllowedMethods cobre o caso específico
// pedido pelo usuário: um preflight OPTIONS anunciando
// Access-Control-Request-Method: PATCH precisa receber PATCH de volta em
// Access-Control-Allow-Methods, senão o browser nem chega a mandar a
// requisição real de PATCH /users/me/voice-preference.
func TestCORS_Preflight_PatchIsInAllowedMethods(t *testing.T) {
	router := newTestCORSRouter(false)

	req := httptest.NewRequest(http.MethodOptions, "/probe", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "PATCH")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("esperava Access-Control-Allow-Origin no preflight, veio %q", got)
	}
	allowMethods := rec.Header().Get("Access-Control-Allow-Methods")
	if !containsToken(allowMethods, "PATCH") {
		t.Fatalf("esperava PATCH em Access-Control-Allow-Methods, veio %q", allowMethods)
	}
}

// TestCORS_DisallowedOrigin_NeverGetsAllowOrigin confirma que a correção do
// bug de PATCH não afrouxou a whitelist de origin — uma origin fora da
// lista continua sem header, pra qualquer método.
func TestCORS_DisallowedOrigin_NeverGetsAllowOrigin(t *testing.T) {
	router := newTestCORSRouter(false)

	req := httptest.NewRequest(http.MethodPatch, "/probe", nil)
	req.Header.Set("Origin", "http://evil.example.com")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("esperava sem Access-Control-Allow-Origin pra origin não permitida, veio %q", got)
	}
}

// TestCORS_LocalNetwork_DisabledByDefault confirma que sem
// CORS_ALLOW_LOCAL_NETWORK (allowLocalNetwork=false) uma origin de rede
// privada continua bloqueada — a feature é aditiva e opt-in, nunca muda o
// comportamento default.
func TestCORS_LocalNetwork_DisabledByDefault(t *testing.T) {
	router := newTestCORSRouter(false)

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("Origin", "http://192.168.1.42:5173")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("esperava sem Access-Control-Allow-Origin pra IP de rede privada com a feature desligada, veio %q", got)
	}
}

// TestCORS_LocalNetwork_AllowsPrivateIPs cobre o pedido do usuário: com
// CORS_ALLOW_LOCAL_NETWORK ligado, qualquer origin http:// com host IP de
// rede privada (as 3 faixas RFC1918) ou loopback é aceita, sem precisar
// listar um IP fixo em CORS_ALLOWED_ORIGINS.
func TestCORS_LocalNetwork_AllowsPrivateIPs(t *testing.T) {
	router := newTestCORSRouter(true)

	for _, origin := range []string{
		"http://192.168.1.42:5173",
		"http://10.0.0.5:5173",
		"http://172.16.4.9:5173",
		"http://127.0.0.1:5173",
	} {
		t.Run(origin, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/probe", nil)
			req.Header.Set("Origin", origin)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != origin {
				t.Fatalf("esperava Access-Control-Allow-Origin=%q, veio %q", origin, got)
			}
		})
	}
}

// TestCORS_LocalNetwork_StillRejectsPublicAndNonIPOrigins garante que
// ligar a feature não abre CORS geral: continua rejeitando origin pública
// (mesmo que a whitelist estática não tenha mudado), HTTPS de IP privado
// (dev local não tem certificado válido pro IP da máquina) e hostname que
// não seja um IP literal (evita abrir pra qualquer domínio arbitrário).
func TestCORS_LocalNetwork_StillRejectsPublicAndNonIPOrigins(t *testing.T) {
	router := newTestCORSRouter(true)

	for _, origin := range []string{
		"http://8.8.8.8:5173",          // IP público
		"https://192.168.1.42:5173",    // IP privado mas HTTPS
		"http://meu-pc.local:5173",     // hostname, não IP literal
		"http://evil.example.com:5173", // origin pública arbitrária
	} {
		t.Run(origin, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/probe", nil)
			req.Header.Set("Origin", origin)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
				t.Fatalf("esperava sem Access-Control-Allow-Origin pra %q, veio %q", origin, got)
			}
		})
	}
}

func containsToken(csv, token string) bool {
	for part := range strings.SplitSeq(csv, ",") {
		if strings.TrimSpace(part) == token {
			return true
		}
	}
	return false
}
