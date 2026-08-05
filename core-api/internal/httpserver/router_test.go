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
func newTestCORSRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(cors.Handler(corsOptions([]string{"http://localhost:5173"})))
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
	router := newTestCORSRouter()

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
	router := newTestCORSRouter()

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
	router := newTestCORSRouter()

	req := httptest.NewRequest(http.MethodPatch, "/probe", nil)
	req.Header.Set("Origin", "http://evil.example.com")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("esperava sem Access-Control-Allow-Origin pra origin não permitida, veio %q", got)
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
