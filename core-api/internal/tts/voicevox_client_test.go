package tts_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yamabiko/core-api/internal/tts"
)

func TestSynthesize_CallsAudioQueryThenSynthesisAndReturnsWav(t *testing.T) {
	var audioQueryCalls, synthesisCalls int
	fakeQuery := `{"accent_phrases":[],"speedScale":1.0}`
	fakeWav := []byte("RIFF....WAVEfake-audio-bytes")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/audio_query":
			audioQueryCalls++
			if r.Method != http.MethodPost {
				t.Fatalf("esperava POST em /audio_query, veio %s", r.Method)
			}
			if got := r.URL.Query().Get("text"); got != "こんにちは" {
				t.Fatalf("esperava text=こんにちは, veio %q", got)
			}
			if got := r.URL.Query().Get("speaker"); got != "1" {
				t.Fatalf("esperava speaker=1, veio %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fakeQuery))
		case "/synthesis":
			synthesisCalls++
			if r.Method != http.MethodPost {
				t.Fatalf("esperava POST em /synthesis, veio %s", r.Method)
			}
			if got := r.URL.Query().Get("speaker"); got != "1" {
				t.Fatalf("esperava speaker=1, veio %q", got)
			}
			body, _ := io.ReadAll(r.Body)
			if string(body) != fakeQuery {
				t.Fatalf("esperava que o body de /synthesis fosse a resposta de /audio_query, veio %q", body)
			}
			w.Header().Set("Content-Type", "audio/wav")
			_, _ = w.Write(fakeWav)
		default:
			t.Fatalf("path inesperado: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := tts.NewVoicevoxClient(server.URL)
	audio, err := client.Synthesize(context.Background(), "こんにちは", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(audio) != string(fakeWav) {
		t.Fatalf("esperava o WAV retornado por /synthesis, veio %q", audio)
	}
	if audioQueryCalls != 1 || synthesisCalls != 1 {
		t.Fatalf("esperava 1 chamada a cada endpoint, veio audio_query=%d synthesis=%d", audioQueryCalls, synthesisCalls)
	}
}

func TestSynthesize_ReturnsErrorWhenAudioQueryFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":"texto inválido"}`))
	}))
	defer server.Close()

	client := tts.NewVoicevoxClient(server.URL)
	_, err := client.Synthesize(context.Background(), "", "1")
	if err == nil {
		t.Fatal("esperava erro quando /audio_query falha")
	}
}

func TestSynthesize_ReturnsErrorWhenSynthesisFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/audio_query" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("engine error"))
	}))
	defer server.Close()

	client := tts.NewVoicevoxClient(server.URL)
	_, err := client.Synthesize(context.Background(), "こんにちは", "1")
	if err == nil {
		t.Fatal("esperava erro quando /synthesis falha")
	}
}
