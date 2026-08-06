package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "email, password e name são obrigatórios")
		return
	}

	tokens, err := h.service.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			writeError(w, http.StatusConflict, "email já cadastrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao registrar usuário")
		return
	}

	writeJSON(w, http.StatusCreated, tokens)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}

	tokens, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "credenciais inválidas")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao autenticar")
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}

	accessToken, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "refresh token inválido")
		return
	}

	writeJSON(w, http.StatusOK, accessTokenResponse{AccessToken: accessToken})
}

// Profiles serve GET /auth/profiles — público, projeção mínima (Sec. 2 do
// pedido do usuário).
func (h *Handler) Profiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.service.Profiles(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar perfis")
		return
	}
	writeJSON(w, http.StatusOK, profiles)
}

type pinLoginRequest struct {
	UserID string `json:"user_id"`
	Pin    string `json:"pin"`
}

// PinLogin serve POST /auth/pin-login (Sec. 3 do pedido do usuário).
func (h *Handler) PinLogin(w http.ResponseWriter, r *http.Request) {
	var req pinLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "user_id inválido")
		return
	}

	tokens, err := h.service.PinLogin(r.Context(), userID, req.Pin, time.Now())
	if err != nil {
		var lockedErr *ErrPinLocked
		if errors.As(err, &lockedErr) {
			writeJSON(w, http.StatusTooManyRequests, map[string]any{
				"error":               "conta temporariamente bloqueada por tentativas de PIN incorretas",
				"retry_after_seconds": int(lockedErr.RetryAfter.Seconds()),
			})
			return
		}
		var invalidErr *ErrPinInvalid
		if errors.As(err, &invalidErr) {
			body := map[string]any{"error": "PIN inválido"}
			if invalidErr.AttemptsRemaining != nil {
				body["attempts_remaining"] = *invalidErr.AttemptsRemaining
			}
			writeJSON(w, http.StatusUnauthorized, body)
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao autenticar")
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

type pinSetupRequest struct {
	Pin string `json:"pin"`
}

var pinFormatPattern = regexp.MustCompile(`^\d{6}$`)

// PinSetup serve POST /auth/pin-setup — autenticado pelo login por senha
// existente (Sec. 4 do pedido do usuário), única porta de entrada pra
// configurar ou resetar o PIN.
func (h *Handler) PinSetup(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "usuário não autenticado")
		return
	}

	var req pinSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if !pinFormatPattern.MatchString(req.Pin) {
		writeError(w, http.StatusBadRequest, "pin precisa ter exatamente 6 dígitos numéricos")
		return
	}

	if err := h.service.SetPin(r.Context(), userID, req.Pin); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao salvar PIN")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
