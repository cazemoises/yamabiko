package auth

import (
	"encoding/json"
	"errors"
	"net/http"
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
