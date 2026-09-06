package handler

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/service"
	"advanced-blog-management-system/pkg/auth"
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	userService *service.UserService
	jwtSecret   string
}

func NewAuthHandler(userService *service.UserService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtSecret:   jwtSecret,
	}
}

// RegisterHandler обрабатывает POST /api/register
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req model.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		h.respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userService.Register(r.Context(), &req)
	if err != nil {
		h.respondWithError(w, err.Error(), apperrors.ToHTTPStatus(err))
		return
	}

	token, expiresAt, err := auth.GenerateToken(user.ID, user.Email, user.Username, h.jwtSecret)
	if err != nil {
		h.respondWithError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, model.TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.ToResponse(),
	})
}

// LoginHandler обрабатывает POST /api/login
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req model.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		h.respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userService.Login(r.Context(), &req)
	if err != nil {
		h.respondWithError(w, err.Error(), apperrors.ToHTTPStatus(err))
		return
	}

	token, expiresAt, err := auth.GenerateToken(user.ID, user.Email, user.Username, h.jwtSecret)
	if err != nil {
		h.respondWithError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	h.respondWithJSON(w, http.StatusOK, model.TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.ToResponse(),
	})
}

// respondWithJSON - helper для отправки JSON ответов
func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// respondWithError - helper для отправки ошибок
func (h *AuthHandler) respondWithError(w http.ResponseWriter, message string, code int) {
	type ErrorResponse struct {
		Error string `json:"error"`
	}
	h.respondWithJSON(w, code, ErrorResponse{Error: message})
}
