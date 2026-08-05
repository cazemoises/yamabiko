package users

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yamabiko/core-api/internal/exercises"
	appmiddleware "github.com/yamabiko/core-api/internal/middleware"
	"github.com/yamabiko/core-api/internal/phonetics"
	"github.com/yamabiko/core-api/internal/srs"
	"github.com/yamabiko/core-api/internal/tts"
)

type Handler struct {
	repo          Repository
	srsRepo       srs.Repository
	phoneticsRepo phonetics.Repository
	exercisesRepo exercises.Repository
}

func NewHandler(repo Repository, srsRepo srs.Repository, phoneticsRepo phonetics.Repository, exercisesRepo exercises.Repository) *Handler {
	return &Handler{repo: repo, srsRepo: srsRepo, phoneticsRepo: phoneticsRepo, exercisesRepo: exercisesRepo}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "usuário não autenticado")
		return
	}

	profile, err := h.repo.FindProfileByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar perfil")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

type progressResponse struct {
	ChunksSolidos         int                      `json:"chunks_solidos"`
	ChunksEmReforco       int                      `json:"chunks_em_reforco"`
	ChunksNovos           int                      `json:"chunks_novos"`
	PhoneticErrorPatterns []phonetics.PatternCount `json:"phonetic_error_patterns"`
}

func (h *Handler) Progress(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "usuário não autenticado")
		return
	}

	counts, err := h.srsRepo.CountByStatus(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar progresso de chunks")
		return
	}

	allExercises, err := h.exercisesRepo.List(r.Context(), exercises.Filter{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar exercícios")
		return
	}

	patterns, err := h.phoneticsRepo.Heatmap(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar padrões fonéticos")
		return
	}

	solido := counts[srs.StatusSolido]
	emReforco := counts[srs.StatusEmReforco]
	novos := max(len(allExercises)-solido-emReforco, 0)

	writeJSON(w, http.StatusOK, progressResponse{
		ChunksSolidos:         solido,
		ChunksEmReforco:       emReforco,
		ChunksNovos:           novos,
		PhoneticErrorPatterns: patterns,
	})
}

type voicePreferenceRequest struct {
	Language string `json:"language"`
	VoiceID  string `json:"voice_id"`
}

// UpdateVoicePreference serve PATCH /users/me/voice-preference. voice_id
// vazio limpa a preferência (volta pro default do idioma em
// tts.GetReferenceAudio); voice_id não-vazio é validado contra o catálogo
// curado (tts.VoicesForLanguage) — só pra language que ele já pertence.
func (h *Handler) UpdateVoicePreference(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "usuário não autenticado")
		return
	}

	var req voicePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if req.Language == "" {
		writeError(w, http.StatusBadRequest, "language é obrigatório")
		return
	}

	if req.VoiceID != "" {
		valid := false
		for _, v := range tts.VoicesForLanguage(req.Language) {
			if v.ID == req.VoiceID {
				valid = true
				break
			}
		}
		if !valid {
			writeError(w, http.StatusBadRequest, "voice_id desconhecido pra esse idioma")
			return
		}
	}

	if err := h.repo.SetVoicePreference(r.Context(), userID, req.Language, req.VoiceID); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			writeError(w, http.StatusNotFound, "usuário não encontrado")
		case errors.Is(err, ErrUnsupportedVoiceLanguage):
			writeError(w, http.StatusBadRequest, "language não suportado pra preferência de voz")
		default:
			writeError(w, http.StatusInternalServerError, "erro ao salvar preferência de voz")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
