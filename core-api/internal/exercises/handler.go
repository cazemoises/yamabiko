package exercises

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	filter := Filter{}
	q := r.URL.Query()

	if v := q.Get("sprint_day"); v != "" {
		day, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "sprint_day inválido")
			return
		}
		filter.SprintDay = &day
	}
	if v := q.Get("category"); v != "" {
		filter.Category = &v
	}
	if v := q.Get("difficulty"); v != "" {
		difficulty, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "difficulty inválido")
			return
		}
		filter.Difficulty = &difficulty
	}

	list, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar exercícios")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	exercise, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrExerciseNotFound) {
			writeError(w, http.StatusNotFound, "exercício não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar exercício")
		return
	}
	writeJSON(w, http.StatusOK, exercise)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
