package handler

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/middleware"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// parsePagination читает limit/offset из query-параметров, приводя их
// к допустимому диапазону: limit по умолчанию 10, максимум 100;
// offset по умолчанию 0, не может быть отрицательным
func parsePagination(r *http.Request) (limit, offset int) {
	limit, offset = 10, 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return limit, offset
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.respondWithError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var req model.PostCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		h.respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}

	post, err := h.postService.CreatePost(r.Context(), &req, userID)
	if err != nil {
		h.respondWithError(w, err.Error(), apperrors.ToHTTPStatus(err))
		return
	}

	h.respondWithJSON(w, http.StatusCreated, post)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		h.respondWithError(w, "invalid post id", http.StatusBadRequest)
		return
	}

	post, err := h.postService.GetPost(r.Context(), id)
	if err != nil {
		h.respondWithError(w, err.Error(), apperrors.ToHTTPStatus(err))
		return
	}

	h.respondWithJSON(w, http.StatusOK, post)
}

func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)

	posts, err := h.postService.GetAllPosts(r.Context(), limit, offset)
	if err != nil {
		h.respondWithError(w, err.Error(), apperrors.ToHTTPStatus(err))
		return
	}

	total, err := h.postService.GetPostsCount(r.Context())
	if err != nil {
		h.respondWithError(w, err.Error(), apperrors.ToHTTPStatus(err))
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"posts":  posts,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.respondWithError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		h.respondWithError(w, "invalid post id", http.StatusBadRequest)
		return
	}

	var req model.PostUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		h.respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}

	post, err := h.postService.UpdatePost(r.Context(), id, &req, userID)
	if err != nil {
		h.respondWithError(w, err.Error(), apperrors.ToHTTPStatus(err))
		return
	}

	h.respondWithJSON(w, http.StatusOK, post)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.respondWithError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		h.respondWithError(w, "invalid post id", http.StatusBadRequest)
		return
	}

	if err := h.postService.DeletePost(r.Context(), id, userID); err != nil {
		h.respondWithError(w, err.Error(), apperrors.ToHTTPStatus(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PostHandler) GetPostsByAuthor(w http.ResponseWriter, r *http.Request) {
	authorID, err := strconv.Atoi(chi.URLParam(r, "authorID"))
	if err != nil || authorID <= 0 {
		h.respondWithError(w, "invalid author id", http.StatusBadRequest)
		return
	}

	limit, offset := parsePagination(r)

	posts, err := h.postService.GetPostsByAuthor(r.Context(), authorID, limit, offset)
	if err != nil {
		h.respondWithError(w, err.Error(), apperrors.ToHTTPStatus(err))
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"posts":  posts,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *PostHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *PostHandler) respondWithError(w http.ResponseWriter, message string, code int) {
	type ErrorResponse struct {
		Error string `json:"error"`
	}
	h.respondWithJSON(w, code, ErrorResponse{Error: message})
}
