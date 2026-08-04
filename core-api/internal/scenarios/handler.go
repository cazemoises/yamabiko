package scenarios

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/exercises"
)

type Handler struct {
	repo          Repository
	exercisesRepo exercises.Repository
}

func NewHandler(repo Repository, exercisesRepo exercises.Repository) *Handler {
	return &Handler{repo: repo, exercisesRepo: exercisesRepo}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	filter := Filter{}
	if v := r.URL.Query().Get("language"); v != "" {
		filter.Language = &v
	}

	list, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar cenários")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

type detailResponse struct {
	Scenario
	Exercises []exercises.Exercise `json:"exercises"`
}

// Detail serve GET /scenarios/{id} — devolve o cenário junto com seus
// exercícios já ordenados por order_in_scenario, pro frontend navegar a
// sequência inteira sem precisar de N+1 requisições nem voltar pra lista
// entre um exercício e outro.
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id de cenário inválido")
		return
	}

	scenario, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrScenarioNotFound) {
			writeError(w, http.StatusNotFound, "cenário não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar cenário")
		return
	}

	exs, err := h.exercisesRepo.List(r.Context(), exercises.Filter{ScenarioID: &id})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar exercícios do cenário")
		return
	}

	writeJSON(w, http.StatusOK, detailResponse{Scenario: *scenario, Exercises: exs})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
