package tts_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/yamabiko/core-api/internal/tts"
)

func newTestRouter(handler *tts.Handler) http.Handler {
	r := chi.NewRouter()
	r.Route("/tts/voices", func(r chi.Router) {
		r.Get("/", handler.Voices)
		r.Get("/{voice_id}/preview", handler.VoicePreview)
	})
	return r
}

func TestHandler_Voices_FiltersByLanguage(t *testing.T) {
	service := tts.NewService(map[string]tts.TTSClient{}, &fakeExerciseFinder{}, t.TempDir())
	router := newTestRouter(tts.NewHandler(service))

	req := httptest.NewRequest(http.MethodGet, "/tts/voices?language=en-US", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, veio %d (%s)", rec.Code, rec.Body.String())
	}

	var voices []tts.Voice
	if err := json.Unmarshal(rec.Body.Bytes(), &voices); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}
	if len(voices) == 0 {
		t.Fatal("esperava pelo menos 1 voz en-US")
	}
	for _, v := range voices {
		if v.Language != "en-US" {
			t.Fatalf("esperava só vozes en-US, veio %q (%s)", v.Language, v.ID)
		}
	}
}

func TestHandler_Voices_NoLanguageFilter_ReturnsWholeCatalog(t *testing.T) {
	service := tts.NewService(map[string]tts.TTSClient{}, &fakeExerciseFinder{}, t.TempDir())
	router := newTestRouter(tts.NewHandler(service))

	req := httptest.NewRequest(http.MethodGet, "/tts/voices", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, veio %d (%s)", rec.Code, rec.Body.String())
	}

	var voices []tts.Voice
	if err := json.Unmarshal(rec.Body.Bytes(), &voices); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}

	languages := map[string]bool{}
	for _, v := range voices {
		languages[v.Language] = true
	}
	if !languages["ja-JP"] || !languages["en-US"] {
		t.Fatalf("esperava vozes dos dois idiomas sem filtro, veio %v", languages)
	}
}

func TestHandler_VoicePreview_ReturnsAudioAndCaches(t *testing.T) {
	cacheDir := t.TempDir()
	synth := &fakeTTSClient{audio: []byte("preview-wav-bytes")}
	service := tts.NewService(map[string]tts.TTSClient{"ja": synth}, &fakeExerciseFinder{}, cacheDir)
	router := newTestRouter(tts.NewHandler(service))

	req := httptest.NewRequest(http.MethodGet, "/tts/voices/ja-announcer-neutral/preview", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, veio %d (%s)", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "preview-wav-bytes" {
		t.Fatalf("esperava o áudio sintetizado no body, veio %q", rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "audio/wav" {
		t.Fatalf("esperava Content-Type audio/wav, veio %q", got)
	}

	// 2ª requisição deve servir do cache, sem chamar o TTSClient de novo.
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/tts/voices/ja-announcer-neutral/preview", nil))
	if rec2.Code != http.StatusOK || rec2.Body.String() != "preview-wav-bytes" {
		t.Fatalf("esperava cache hit servindo o mesmo áudio, veio status=%d body=%q", rec2.Code, rec2.Body.String())
	}
	if synth.calls != 1 {
		t.Fatalf("esperava 1 chamada ao TTSClient (cache hit na 2ª), veio %d", synth.calls)
	}
}

func TestHandler_VoicePreview_UnknownVoiceID_Returns404(t *testing.T) {
	service := tts.NewService(map[string]tts.TTSClient{}, &fakeExerciseFinder{}, t.TempDir())
	router := newTestRouter(tts.NewHandler(service))

	req := httptest.NewRequest(http.MethodGet, "/tts/voices/voice-id-que-nao-existe/preview", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404, veio %d (%s)", rec.Code, rec.Body.String())
	}
}
