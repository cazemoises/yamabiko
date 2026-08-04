package sttclient_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yamabiko/core-api/internal/sttclient"
)

func TestTranscribe_SendsMultipartAndParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transcribe" {
			t.Fatalf("esperava path /transcribe, veio %s", r.URL.Path)
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("esperava arquivo multipart, erro: %v", err)
		}
		defer file.Close()
		if header.Filename != "attempt.wav" {
			t.Fatalf("esperava filename attempt.wav, veio %s", header.Filename)
		}
		content, _ := io.ReadAll(file)
		if string(content) != "fake-audio" {
			t.Fatalf("esperava conteúdo 'fake-audio', veio %q", content)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"transcript":"こんにちは","language":"ja","confidence":0.97}`))
	}))
	defer server.Close()

	client := sttclient.New(server.URL)
	result, err := client.Transcribe(context.Background(), "attempt.wav", strings.NewReader("fake-audio"), "ja-JP")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Transcript != "こんにちは" {
		t.Fatalf("esperava transcript こんにちは, veio %s", result.Transcript)
	}
	if result.Confidence != 0.97 {
		t.Fatalf("esperava confidence 0.97, veio %v", result.Confidence)
	}
}

func TestTranscribe_SendsLanguageField(t *testing.T) {
	var receivedLanguage string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("esperava multipart form válido, erro: %v", err)
		}
		receivedLanguage = r.FormValue("language")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"transcript":"hi","language":"en","confidence":0.9}`))
	}))
	defer server.Close()

	client := sttclient.New(server.URL)
	if _, err := client.Transcribe(context.Background(), "attempt.webm", strings.NewReader("fake-audio"), "en-US"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedLanguage != "en-US" {
		t.Fatalf("esperava campo language='en-US' no multipart, veio %q", receivedLanguage)
	}
}

func TestTranscribe_ReturnsErrorOnNonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":"formato inválido"}`))
	}))
	defer server.Close()

	client := sttclient.New(server.URL)
	_, err := client.Transcribe(context.Background(), "attempt.wav", strings.NewReader("fake-audio"), "ja-JP")
	if err == nil {
		t.Fatal("esperava erro pra status não-OK")
	}
}
