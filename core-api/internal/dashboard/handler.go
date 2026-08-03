// Package dashboard expõe agregações de progresso pro frontend (Sec. 4 do
// CLAUDE.md — GET /dashboard/heatmap).
package dashboard

import (
	"encoding/json"
	"net/http"

	appmiddleware "github.com/yamabiko/core-api/internal/middleware"
	"github.com/yamabiko/core-api/internal/phonetics"
)

type Handler struct {
	phoneticsRepo phonetics.Repository
}

func NewHandler(phoneticsRepo phonetics.Repository) *Handler {
	return &Handler{phoneticsRepo: phoneticsRepo}
}

func (h *Handler) Heatmap(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "usuário não autenticado")
		return
	}

	patterns, err := h.phoneticsRepo.Heatmap(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar heatmap")
		return
	}
	writeJSON(w, http.StatusOK, patterns)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
