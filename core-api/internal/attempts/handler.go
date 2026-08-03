package attempts

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/exercises"
	appmiddleware "github.com/yamabiko/core-api/internal/middleware"
)

const maxUploadSize = 20 << 20 // 20MB

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type submitResponse struct {
	Transcript   string   `json:"transcript"`
	Score        float64  `json:"score"`
	Verdict      string   `json:"verdict"`
	Diff         any      `json:"diff"`
	XPGained     int      `json:"xp_gained"`
	BadgesGained []string `json:"badges_gained,omitempty"`
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "usuário não autenticado")
		return
	}

	exerciseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id de exercício inválido")
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "corpo multipart inválido")
		return
	}
	file, header, err := r.FormFile("audio")
	if err != nil {
		writeError(w, http.StatusBadRequest, "arquivo 'audio' é obrigatório")
		return
	}
	defer file.Close()

	result, err := h.service.Submit(r.Context(), userID, exerciseID, header.Filename, file)
	if err != nil {
		if errors.Is(err, exercises.ErrExerciseNotFound) {
			writeError(w, http.StatusNotFound, "exercício não encontrado")
			return
		}
		writeError(w, http.StatusBadGateway, "erro ao processar tentativa: "+err.Error())
		return
	}

	badges := make([]string, len(result.NewBadges))
	for i, badge := range result.NewBadges {
		badges[i] = string(badge)
	}

	writeJSON(w, http.StatusCreated, submitResponse{
		Transcript:   result.Transcript,
		Score:        result.Score,
		Verdict:      string(result.Verdict),
		Diff:         result.Diff,
		XPGained:     result.XPGained,
		BadgesGained: badges,
	})
}

func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "usuário não autenticado")
		return
	}

	exerciseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id de exercício inválido")
		return
	}

	history, err := h.service.History(r.Context(), userID, exerciseID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar histórico")
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
