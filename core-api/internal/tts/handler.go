package tts

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/exercises"
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

	audio, err := h.service.GetReferenceAudio(r.Context(), id)
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

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
