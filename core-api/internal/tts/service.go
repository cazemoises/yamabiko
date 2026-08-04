package tts

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/exercises"
)

// ErrLanguageNotSupported é devolvido quando o exercício não é ja-JP — o
// VOICEVOX só sintetiza japonês, então pra outros idiomas (ex: en-US) o
// frontend deve usar a Web Speech API do próprio browser em vez deste
// endpoint (Sec. pedida pelo usuário).
var ErrLanguageNotSupported = errors.New("tts: idioma do exercício não é suportado pelo VOICEVOX")

type Synthesizer interface {
	Synthesize(ctx context.Context, text string) ([]byte, error)
}

type ExerciseFinder interface {
	FindByID(ctx context.Context, id uuid.UUID) (*exercises.Exercise, error)
}

// Service gera (e cacheia em disco) o áudio de referência em japonês de um
// exercício. O cache existe porque expected_transcript é estático por
// exercício — não faz sentido chamar o VOICEVOX de novo a cada request pro
// mesmo exercício (mesmo raciocínio de custo/latência do débito documentado
// em BUILD_STATE.md sobre a Web Speech API, mas resolvido de fato aqui).
type Service struct {
	synthesizer    Synthesizer
	exerciseFinder ExerciseFinder
	cacheDir       string
}

func NewService(synthesizer Synthesizer, exerciseFinder ExerciseFinder, cacheDir string) *Service {
	return &Service{synthesizer: synthesizer, exerciseFinder: exerciseFinder, cacheDir: cacheDir}
}

func (s *Service) GetReferenceAudio(ctx context.Context, exerciseID uuid.UUID) ([]byte, error) {
	exercise, err := s.exerciseFinder.FindByID(ctx, exerciseID)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(strings.ToLower(exercise.Language), "ja") {
		return nil, ErrLanguageNotSupported
	}

	cachePath := s.cachePath(exerciseID)
	if cached, err := os.ReadFile(cachePath); err == nil {
		return cached, nil
	}

	audio, err := s.synthesizer.Synthesize(ctx, exercise.ExpectedTranscript)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(s.cacheDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(cachePath, audio, 0o644); err != nil {
		return nil, err
	}

	return audio, nil
}

func (s *Service) cachePath(exerciseID uuid.UUID) string {
	return filepath.Join(s.cacheDir, exerciseID.String()+".wav")
}
