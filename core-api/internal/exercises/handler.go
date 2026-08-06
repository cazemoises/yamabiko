package exercises

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/comparison"
	"github.com/yamabiko/core-api/internal/exercises/validation"
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
	if v := q.Get("language"); v != "" {
		filter.Language = &v
	}
	if v := q.Get("scenario_id"); v != "" {
		scenarioID, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "scenario_id inválido")
			return
		}
		filter.ScenarioID = &scenarioID
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

// Answer serve POST /exercises/{id}/answer — os 5 tipos binários
// certo/errado (multiple_choice_translation, word_order, verb_conjugation,
// matching_pairs, true_false). Não toca no fluxo de áudio existente
// (attempts/comparison/stt-service): é uma validação stateless contra
// type_data, sem persistir tentativa nem tocar XP/streak/SRS (fora do
// escopo pedido — ver BUILD_STATE.md).
func (h *Handler) Answer(w http.ResponseWriter, r *http.Request) {
	exercise, ok := h.findExerciseOrWriteError(w, r)
	if !ok {
		return
	}

	var req validation.AnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	result, err := validation.ValidateAnswer(exercise.ExerciseType, exercise.TypeData, req)
	if err != nil {
		switch {
		case errors.Is(err, validation.ErrUnsupportedType):
			writeError(w, http.StatusBadRequest, "exercise_type '"+exercise.ExerciseType+"' não usa POST /answer — use /text-attempt (dictation, free_translation) ou /attempts (audio_pronunciation)")
		case errors.Is(err, validation.ErrMissingAnswerField):
			writeError(w, http.StatusBadRequest, "campo de resposta ausente ou no formato errado pro exercise_type '"+exercise.ExerciseType+"'")
		case errors.Is(err, validation.ErrInvalidTypeData):
			writeError(w, http.StatusInternalServerError, "type_data do exercício está mal formado")
		default:
			writeError(w, http.StatusInternalServerError, "erro ao validar resposta")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

type textAttemptRequest struct {
	Transcript string `json:"transcript"`
}

type textAttemptResponse struct {
	Transcript string                 `json:"transcript"`
	// Expected é o texto contra o qual o backend de fato comparou —
	// exercise.ExpectedTranscript pra dictation, mas pra free_translation é
	// QUAL das várias acceptable_answers venceu (o frontend não tem como
	// adivinhar isso só com type_data, que lista todas as opções).
	Expected string                 `json:"expected"`
	Score    float64                `json:"score"`
	Verdict  comparison.Verdict     `json:"verdict"`
	Diff     []comparison.DiffEntry `json:"diff"`
}

// TextAttempt serve POST /exercises/{id}/text-attempt — dictation e
// free_translation, os 2 tipos que reaproveitam comparison/ (Levenshtein)
// texto-a-texto em vez do fluxo binário de Answer. Não passa pelo
// stt-service (o aluno digita, não grava áudio) nem pelo fluxo de
// attempts/gamification/SRS do áudio — mesma decisão de escopo de Answer.
func (h *Handler) TextAttempt(w http.ResponseWriter, r *http.Request) {
	exercise, ok := h.findExerciseOrWriteError(w, r)
	if !ok {
		return
	}

	var req textAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	var result comparison.Result
	var expected string
	switch exercise.ExerciseType {
	case "dictation":
		expected = exercise.ExpectedTranscript
		result = validation.ValidateDictation(expected, exercise.Language, req.Transcript)
	case "free_translation":
		var err error
		result, expected, err = validation.ValidateFreeTranslation(exercise.TypeData, exercise.Language, req.Transcript)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "type_data do exercício está mal formado")
			return
		}
	default:
		writeError(w, http.StatusBadRequest, "exercise_type '"+exercise.ExerciseType+"' não usa POST /text-attempt — use /answer (tipos de escolha) ou /attempts (audio_pronunciation)")
		return
	}

	writeJSON(w, http.StatusOK, textAttemptResponse{
		Transcript: req.Transcript,
		Expected:   expected,
		Score:      result.SimilarityScore,
		Verdict:    result.Verdict,
		Diff:       result.PhoneticDiff,
	})
}

// findExerciseOrWriteError é o preâmbulo compartilhado por Answer e
// TextAttempt: parse do {id} da URL + busca no repositório, já escrevendo a
// resposta de erro certa (400/404/500) e devolvendo ok=false quando não dá
// pra continuar.
func (h *Handler) findExerciseOrWriteError(w http.ResponseWriter, r *http.Request) (*Exercise, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return nil, false
	}

	exercise, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrExerciseNotFound) {
			writeError(w, http.StatusNotFound, "exercício não encontrado")
			return nil, false
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar exercício")
		return nil, false
	}
	return exercise, true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
