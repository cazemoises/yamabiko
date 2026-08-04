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

// ErrLanguageNotSupported é devolvido quando não há TTSClient registrado pro
// idioma do exercício — hoje o produto cobre ja-JP (VOICEVOX) e en-US
// (Piper), então isso só aconteceria pra um idioma futuro sem motor de TTS
// configurado ainda.
var ErrLanguageNotSupported = errors.New("tts: nenhum motor de TTS configurado pro idioma desse exercício")

type ExerciseFinder interface {
	FindByID(ctx context.Context, id uuid.UUID) (*exercises.Exercise, error)
}

// Service gera (e cacheia em disco) o áudio de referência de um exercício,
// escolhendo o TTSClient certo pelo idioma do exercício (chave do mapa:
// subtag primário do idioma, ex: "ja" de "ja-JP", "en" de "en-US"). O cache
// existe porque expected_transcript é estático por exercício — não faz
// sentido chamar o motor de TTS de novo a cada request pro mesmo exercício
// (mesmo raciocínio de custo/latência do débito documentado em
// BUILD_STATE.md sobre a Web Speech API, mas resolvido de fato aqui). A
// chave de cache é só o exercise_id, então funciona pros dois idiomas sem
// mudança — cada exercício só tem 1 idioma, nunca colide.
type Service struct {
	clients        map[string]TTSClient
	exerciseFinder ExerciseFinder
	cacheDir       string
}

func NewService(clients map[string]TTSClient, exerciseFinder ExerciseFinder, cacheDir string) *Service {
	return &Service{clients: clients, exerciseFinder: exerciseFinder, cacheDir: cacheDir}
}

func (s *Service) GetReferenceAudio(ctx context.Context, exerciseID uuid.UUID) ([]byte, error) {
	exercise, err := s.exerciseFinder.FindByID(ctx, exerciseID)
	if err != nil {
		return nil, err
	}

	client, ok := s.clients[primaryLanguageSubtag(exercise.Language)]
	if !ok {
		return nil, ErrLanguageNotSupported
	}

	cachePath := s.cachePath(exerciseID)
	if cached, err := os.ReadFile(cachePath); err == nil {
		return cached, nil
	}

	audio, err := client.Synthesize(ctx, exercise.ExpectedTranscript)
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

// primaryLanguageSubtag extrai o subtag primário de uma tag BCP-47 (ex: "ja"
// de "ja-JP", "en" de "en-US") — é a chave usada tanto aqui quanto em
// comparison.CompareLang pra despachar por idioma sem depender da região.
func primaryLanguageSubtag(language string) string {
	return strings.ToLower(strings.SplitN(language, "-", 2)[0])
}
