package tts

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/exercises"
)

// ErrLanguageNotSupported é devolvido quando não há TTSClient registrado pro
// idioma do exercício — hoje o produto cobre ja-JP (VOICEVOX) e en-US
// (Piper), então isso só aconteceria pra um idioma futuro sem motor de TTS
// configurado ainda.
var ErrLanguageNotSupported = errors.New("tts: nenhum motor de TTS configurado pro idioma desse exercício")

// ErrVoiceNotFound é devolvido por GetVoicePreview quando o voice_id não
// existe no catálogo curado (voice.go). GetReferenceAudio nunca devolve
// esse erro — um voice_id desconhecido ou de idioma incompatível ali cai
// no default do idioma do exercício (fail-open), porque é sempre possível
// tocar áudio de referência mesmo que a preferência salva do usuário tenha
// ficado obsoleta.
var ErrVoiceNotFound = errors.New("tts: voice_id desconhecido")

// previewTexts é o texto curto sintetizado por GetVoicePreview pra cada
// idioma — só precisa dar uma amostra representativa da voz, não é
// conteúdo de exercício.
var previewTexts = map[string]string{
	"ja": "こんにちは、よろしくお願いします。",
	"en": "Hello, nice to meet you.",
}

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

// GetReferenceAudio sintetiza (ou serve do cache) o expected_transcript do
// exercício na voz indicada. voiceID é o id estável do catálogo
// (voice.go) — "" usa o default do idioma do exercício, e um voiceID
// desconhecido ou de outro idioma também cai no default (fail-open: a
// preferência do usuário pode ter ficado obsoleta se o catálogo mudou, mas
// isso nunca deveria impedir o áudio de referência de tocar).
func (s *Service) GetReferenceAudio(ctx context.Context, exerciseID uuid.UUID, voiceID string) ([]byte, error) {
	exercise, err := s.exerciseFinder.FindByID(ctx, exerciseID)
	if err != nil {
		return nil, err
	}

	voice, ok := s.resolveVoice(voiceID, exercise.Language)
	if !ok {
		return nil, ErrLanguageNotSupported
	}

	client, ok := s.clients[primaryLanguageSubtag(voice.Language)]
	if !ok {
		return nil, ErrLanguageNotSupported
	}

	cachePath := s.referenceCachePath(exerciseID, voice.ID)
	if cached, err := os.ReadFile(cachePath); err == nil {
		return cached, nil
	}

	audio, err := client.Synthesize(ctx, exercise.ExpectedTranscript, voice.providerVoiceID)
	if err != nil {
		return nil, err
	}

	if err := s.writeCache(cachePath, audio); err != nil {
		return nil, err
	}

	return audio, nil
}

// GetVoicePreview sintetiza (ou serve do cache) uma frase curta na voz
// indicada, cacheada por voice_id — ao contrário de GetReferenceAudio, um
// voice_id desconhecido aqui é erro de verdade (ErrVoiceNotFound): não há
// texto de exercício nenhum pra cair em fallback, o pedido é
// especificamente "toca essa voz".
func (s *Service) GetVoicePreview(ctx context.Context, voiceID string) ([]byte, error) {
	voice, ok := findVoice(voiceID)
	if !ok {
		return nil, ErrVoiceNotFound
	}

	client, ok := s.clients[primaryLanguageSubtag(voice.Language)]
	if !ok {
		return nil, ErrLanguageNotSupported
	}

	text, ok := previewTexts[primaryLanguageSubtag(voice.Language)]
	if !ok {
		return nil, ErrLanguageNotSupported
	}

	cachePath := s.previewCachePath(voice.ID)
	if cached, err := os.ReadFile(cachePath); err == nil {
		return cached, nil
	}

	audio, err := client.Synthesize(ctx, text, voice.providerVoiceID)
	if err != nil {
		return nil, err
	}

	if err := s.writeCache(cachePath, audio); err != nil {
		return nil, err
	}

	return audio, nil
}

// resolveVoice resolve voiceID pra sua entrada de catálogo, caindo no
// default do idioma quando voiceID é vazio, desconhecido, ou pertence a um
// idioma diferente do pedido — ok=false só quando nem o idioma pedido tem
// nenhuma voz no catálogo (ErrLanguageNotSupported).
func (s *Service) resolveVoice(voiceID, language string) (Voice, bool) {
	if voiceID != "" {
		if voice, ok := findVoice(voiceID); ok && primaryLanguageSubtag(voice.Language) == primaryLanguageSubtag(language) {
			return voice, true
		}
	}

	defaultID := DefaultVoiceID(language)
	if defaultID == "" {
		return Voice{}, false
	}
	return findVoice(defaultID)
}

func (s *Service) writeCache(cachePath string, audio []byte) error {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(cachePath, audio, 0o644)
}

func (s *Service) referenceCachePath(exerciseID uuid.UUID, voiceID string) string {
	return filepath.Join(s.cacheDir, "exercises", exerciseID.String()+"__"+voiceID+".wav")
}

func (s *Service) previewCachePath(voiceID string) string {
	return filepath.Join(s.cacheDir, "previews", voiceID+".wav")
}
