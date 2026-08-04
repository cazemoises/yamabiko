package tts_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/exercises"
	"github.com/yamabiko/core-api/internal/tts"
)

type fakeTTSClient struct {
	calls int
	audio []byte
	err   error
}

func (f *fakeTTSClient) Synthesize(_ context.Context, _ string) ([]byte, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.audio, nil
}

type fakeExerciseFinder struct {
	exercise *exercises.Exercise
	err      error
}

func (f *fakeExerciseFinder) FindByID(_ context.Context, _ uuid.UUID) (*exercises.Exercise, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.exercise, nil
}

func TestGetReferenceAudio_CacheMiss_SynthesizesAndCaches(t *testing.T) {
	cacheDir := t.TempDir()
	synth := &fakeTTSClient{audio: []byte("fake-wav-bytes")}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{Language: "ja-JP", ExpectedTranscript: "こんにちは"}}
	service := tts.NewService(map[string]tts.TTSClient{"ja": synth}, finder, cacheDir)

	exerciseID := uuid.New()
	audio, err := service.GetReferenceAudio(context.Background(), exerciseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(audio) != "fake-wav-bytes" {
		t.Fatalf("esperava o áudio sintetizado, veio %q", audio)
	}
	if synth.calls != 1 {
		t.Fatalf("esperava 1 chamada ao TTSClient, veio %d", synth.calls)
	}

	cachedPath := filepath.Join(cacheDir, exerciseID.String()+".wav")
	cached, err := os.ReadFile(cachedPath)
	if err != nil {
		t.Fatalf("esperava arquivo cacheado em %s, erro: %v", cachedPath, err)
	}
	if string(cached) != "fake-wav-bytes" {
		t.Fatalf("esperava conteúdo cacheado igual ao áudio sintetizado, veio %q", cached)
	}
}

func TestGetReferenceAudio_CacheHit_DoesNotCallTTSClientAgain(t *testing.T) {
	cacheDir := t.TempDir()
	synth := &fakeTTSClient{audio: []byte("fake-wav-bytes")}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{Language: "ja-JP", ExpectedTranscript: "こんにちは"}}
	service := tts.NewService(map[string]tts.TTSClient{"ja": synth}, finder, cacheDir)

	exerciseID := uuid.New()

	if _, err := service.GetReferenceAudio(context.Background(), exerciseID); err != nil {
		t.Fatalf("unexpected error na 1ª chamada: %v", err)
	}
	if synth.calls != 1 {
		t.Fatalf("esperava 1 chamada após a 1ª requisição, veio %d", synth.calls)
	}

	audio, err := service.GetReferenceAudio(context.Background(), exerciseID)
	if err != nil {
		t.Fatalf("unexpected error na 2ª chamada: %v", err)
	}
	if string(audio) != "fake-wav-bytes" {
		t.Fatalf("esperava o áudio cacheado, veio %q", audio)
	}
	if synth.calls != 1 {
		t.Fatalf("esperava que o TTSClient NÃO fosse chamado de novo (servir do cache), veio %d chamadas", synth.calls)
	}
}

func TestGetReferenceAudio_PreExistingCacheFile_NeverCallsTTSClient(t *testing.T) {
	cacheDir := t.TempDir()
	synth := &fakeTTSClient{audio: []byte("nao-deveria-ser-usado")}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{Language: "ja-JP", ExpectedTranscript: "こんにちは"}}
	service := tts.NewService(map[string]tts.TTSClient{"ja": synth}, finder, cacheDir)

	exerciseID := uuid.New()
	preExisting := []byte("audio-ja-cacheado-antes")
	if err := os.WriteFile(filepath.Join(cacheDir, exerciseID.String()+".wav"), preExisting, 0o644); err != nil {
		t.Fatalf("falha ao preparar cache: %v", err)
	}

	audio, err := service.GetReferenceAudio(context.Background(), exerciseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(audio) != string(preExisting) {
		t.Fatalf("esperava servir o arquivo pré-existente, veio %q", audio)
	}
	if synth.calls != 0 {
		t.Fatalf("esperava 0 chamadas ao TTSClient com cache pré-existente, veio %d", synth.calls)
	}
}

func TestGetReferenceAudio_LanguageWithoutRegisteredClient_ReturnsLanguageNotSupported(t *testing.T) {
	cacheDir := t.TempDir()
	jaClient := &fakeTTSClient{audio: []byte("audio-ja")}
	enClient := &fakeTTSClient{audio: []byte("audio-en")}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{Language: "fr-FR", ExpectedTranscript: "bonjour"}}
	service := tts.NewService(map[string]tts.TTSClient{"ja": jaClient, "en": enClient}, finder, cacheDir)

	_, err := service.GetReferenceAudio(context.Background(), uuid.New())
	if err != tts.ErrLanguageNotSupported {
		t.Fatalf("esperava ErrLanguageNotSupported, veio %v", err)
	}
	if jaClient.calls != 0 || enClient.calls != 0 {
		t.Fatalf("esperava 0 chamadas a qualquer TTSClient pra idioma sem client registrado, veio ja=%d en=%d", jaClient.calls, enClient.calls)
	}
}

func TestGetReferenceAudio_RoutesToTheClientMatchingExerciseLanguage(t *testing.T) {
	cacheDir := t.TempDir()
	jaClient := &fakeTTSClient{audio: []byte("audio-ja")}
	enClient := &fakeTTSClient{audio: []byte("audio-en")}
	clients := map[string]tts.TTSClient{"ja": jaClient, "en": enClient}

	jaFinder := &fakeExerciseFinder{exercise: &exercises.Exercise{Language: "ja-JP", ExpectedTranscript: "こんにちは"}}
	jaService := tts.NewService(clients, jaFinder, cacheDir)
	jaAudio, err := jaService.GetReferenceAudio(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error (ja): %v", err)
	}
	if string(jaAudio) != "audio-ja" || jaClient.calls != 1 || enClient.calls != 0 {
		t.Fatalf("esperava roteamento pro TTSClient 'ja', veio audio=%q ja.calls=%d en.calls=%d", jaAudio, jaClient.calls, enClient.calls)
	}

	enFinder := &fakeExerciseFinder{exercise: &exercises.Exercise{Language: "en-US", ExpectedTranscript: "hello"}}
	enService := tts.NewService(clients, enFinder, cacheDir)
	enAudio, err := enService.GetReferenceAudio(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error (en): %v", err)
	}
	if string(enAudio) != "audio-en" || jaClient.calls != 1 || enClient.calls != 1 {
		t.Fatalf("esperava roteamento pro TTSClient 'en', veio audio=%q ja.calls=%d en.calls=%d", enAudio, jaClient.calls, enClient.calls)
	}
}

func TestGetReferenceAudio_ExerciseNotFound_PropagatesError(t *testing.T) {
	cacheDir := t.TempDir()
	synth := &fakeTTSClient{audio: []byte("fake-wav-bytes")}
	finder := &fakeExerciseFinder{err: exercises.ErrExerciseNotFound}
	service := tts.NewService(map[string]tts.TTSClient{"ja": synth}, finder, cacheDir)

	_, err := service.GetReferenceAudio(context.Background(), uuid.New())
	if err != exercises.ErrExerciseNotFound {
		t.Fatalf("esperava ErrExerciseNotFound, veio %v", err)
	}
}

func TestGetReferenceAudio_TTSClientError_DoesNotCache(t *testing.T) {
	cacheDir := t.TempDir()
	synth := &fakeTTSClient{err: context.DeadlineExceeded}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{Language: "ja-JP", ExpectedTranscript: "こんにちは"}}
	service := tts.NewService(map[string]tts.TTSClient{"ja": synth}, finder, cacheDir)

	exerciseID := uuid.New()
	if _, err := service.GetReferenceAudio(context.Background(), exerciseID); err == nil {
		t.Fatal("esperava erro quando o TTSClient falha")
	}

	if _, err := os.Stat(filepath.Join(cacheDir, exerciseID.String()+".wav")); !os.IsNotExist(err) {
		t.Fatal("não deveria ter criado arquivo de cache quando a síntese falha")
	}
}
