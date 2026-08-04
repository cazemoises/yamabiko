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

type fakeSynthesizer struct {
	calls int
	audio []byte
	err   error
}

func (f *fakeSynthesizer) Synthesize(_ context.Context, _ string) ([]byte, error) {
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
	synth := &fakeSynthesizer{audio: []byte("fake-wav-bytes")}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{Language: "ja-JP", ExpectedTranscript: "こんにちは"}}
	service := tts.NewService(synth, finder, cacheDir)

	exerciseID := uuid.New()
	audio, err := service.GetReferenceAudio(context.Background(), exerciseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(audio) != "fake-wav-bytes" {
		t.Fatalf("esperava o áudio sintetizado, veio %q", audio)
	}
	if synth.calls != 1 {
		t.Fatalf("esperava 1 chamada ao synthesizer, veio %d", synth.calls)
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

func TestGetReferenceAudio_CacheHit_DoesNotCallSynthesizerAgain(t *testing.T) {
	cacheDir := t.TempDir()
	synth := &fakeSynthesizer{audio: []byte("fake-wav-bytes")}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{Language: "ja-JP", ExpectedTranscript: "こんにちは"}}
	service := tts.NewService(synth, finder, cacheDir)

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
		t.Fatalf("esperava que o VOICEVOX NÃO fosse chamado de novo (servir do cache), veio %d chamadas", synth.calls)
	}
}

func TestGetReferenceAudio_PreExistingCacheFile_NeverCallsSynthesizer(t *testing.T) {
	cacheDir := t.TempDir()
	synth := &fakeSynthesizer{audio: []byte("nao-deveria-ser-usado")}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{Language: "ja-JP", ExpectedTranscript: "こんにちは"}}
	service := tts.NewService(synth, finder, cacheDir)

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
		t.Fatalf("esperava 0 chamadas ao synthesizer com cache pré-existente, veio %d", synth.calls)
	}
}

func TestGetReferenceAudio_NonJapaneseExercise_ReturnsLanguageNotSupported(t *testing.T) {
	cacheDir := t.TempDir()
	synth := &fakeSynthesizer{audio: []byte("fake-wav-bytes")}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{Language: "en-US", ExpectedTranscript: "hi, how are you?"}}
	service := tts.NewService(synth, finder, cacheDir)

	_, err := service.GetReferenceAudio(context.Background(), uuid.New())
	if err != tts.ErrLanguageNotSupported {
		t.Fatalf("esperava ErrLanguageNotSupported, veio %v", err)
	}
	if synth.calls != 0 {
		t.Fatalf("esperava 0 chamadas ao synthesizer pra idioma não suportado, veio %d", synth.calls)
	}
}

func TestGetReferenceAudio_ExerciseNotFound_PropagatesError(t *testing.T) {
	cacheDir := t.TempDir()
	synth := &fakeSynthesizer{audio: []byte("fake-wav-bytes")}
	finder := &fakeExerciseFinder{err: exercises.ErrExerciseNotFound}
	service := tts.NewService(synth, finder, cacheDir)

	_, err := service.GetReferenceAudio(context.Background(), uuid.New())
	if err != exercises.ErrExerciseNotFound {
		t.Fatalf("esperava ErrExerciseNotFound, veio %v", err)
	}
}

func TestGetReferenceAudio_SynthesizerError_DoesNotCache(t *testing.T) {
	cacheDir := t.TempDir()
	synth := &fakeSynthesizer{err: context.DeadlineExceeded}
	finder := &fakeExerciseFinder{exercise: &exercises.Exercise{Language: "ja-JP", ExpectedTranscript: "こんにちは"}}
	service := tts.NewService(synth, finder, cacheDir)

	exerciseID := uuid.New()
	if _, err := service.GetReferenceAudio(context.Background(), exerciseID); err == nil {
		t.Fatal("esperava erro quando o synthesizer falha")
	}

	if _, err := os.Stat(filepath.Join(cacheDir, exerciseID.String()+".wav")); !os.IsNotExist(err) {
		t.Fatal("não deveria ter criado arquivo de cache quando a síntese falha")
	}
}
