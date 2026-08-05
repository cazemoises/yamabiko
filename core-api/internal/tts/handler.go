package tts

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/exercises"
	appmiddleware "github.com/yamabiko/core-api/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ReferenceAudio serve GET /exercises/{id}/reference-audio. Funciona pra
// qualquer idioma que tenha um TTSClient registrado no Service (hoje:
// ja-JP via VOICEVOX, en-US via Piper) — pra qualquer outro devolve 404 com
// uma mensagem explícita de que não há motor de TTS configurado.
func (h *Handler) ReferenceAudio(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id de exercício inválido")
		return
	}

	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "usuário não autenticado")
		return
	}

	audio, err := h.service.ReferenceAudioForUser(r.Context(), id, userID)
	switch {
	case errors.Is(err, exercises.ErrExerciseNotFound):
		writeError(w, http.StatusNotFound, "exercício não encontrado")
		return
	case errors.Is(err, ErrLanguageNotSupported):
		writeError(w, http.StatusNotFound, "áudio de referência não disponível pra esse idioma")
		return
	case err != nil:
		writeError(w, http.StatusBadGateway, "erro ao gerar áudio de referência: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "audio/wav")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(audio)
}

// Voices serve GET /tts/voices?language= — devolve o catálogo curado
// (voice.go) do idioma pedido. Sem language= devolve o catálogo inteiro
// (todos os idiomas), útil pro frontend montar o seletor de voz sem saber
// de antemão em qual idioma o usuário está.
func (h *Handler) Voices(w http.ResponseWriter, r *http.Request) {
	language := r.URL.Query().Get("language")

	var voices []Voice
	if language == "" {
		voices = allVoices()
	} else {
		voices = VoicesForLanguage(language)
	}

	writeJSON(w, http.StatusOK, voices)
}

// VoicePreview serve GET /tts/voices/{voice_id}/preview — sintetiza (ou
// serve do cache) uma frase curta na voz pedida, pro usuário ouvir antes de
// escolher como preferência.
func (h *Handler) VoicePreview(w http.ResponseWriter, r *http.Request) {
	voiceID := chi.URLParam(r, "voice_id")

	audio, err := h.service.GetVoicePreview(r.Context(), voiceID)
	switch {
	case errors.Is(err, ErrVoiceNotFound):
		writeError(w, http.StatusNotFound, "voz desconhecida")
		return
	case errors.Is(err, ErrLanguageNotSupported):
		writeError(w, http.StatusNotFound, "preview não disponível pra essa voz")
		return
	case err != nil:
		writeError(w, http.StatusBadGateway, "erro ao gerar preview: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "audio/wav")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(audio)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
