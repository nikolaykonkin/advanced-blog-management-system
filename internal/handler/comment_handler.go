package handler

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/middleware"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// commentServiceErrorStatus расширяет apperrors.ToHTTPStatus для service.ErrPostNotPublished —
// эта ошибка объявлена локально в пакете service, а не в фиксированном списке apperrors, поэтому
// ToHTTPStatus про неё не знает и без этой проверки вернул бы 500 вместо
// корректного 400 ("пост существует, но комментировать его пока нельзя").
func commentServiceErrorStatus(err error) int {
	if errors.Is(err, service.ErrPostNotPublished) {
		return http.StatusBadRequest
	}
	return apperrors.ToHTTPStatus(err)
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.respondWithError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	postID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || postID <= 0 {
		h.respondWithError(w, "invalid post id", http.StatusBadRequest)
		return
	}

	var req model.CommentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		h.respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}

	comment, err := h.commentService.CreateComment(r.Context(), &req, postID, userID)
	if err != nil {
		h.respondWithError(w, err.Error(), commentServiceErrorStatus(err))
		return
	}

	h.respondWithJSON(w, http.StatusCreated, comment)
}

func (h *CommentHandler) GetCommentsByPostID(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || postID <= 0 {
		h.respondWithError(w, "invalid post id", http.StatusBadRequest)
		return
	}

	limit, offset := parsePagination(r)

	comments, err := h.commentService.GetCommentsByPostID(r.Context(), postID, limit, offset)
	if err != nil {
		h.respondWithError(w, err.Error(), commentServiceErrorStatus(err))
		return
	}

	total, err := h.commentService.GetCommentsCountByPostID(r.Context(), postID)
	if err != nil {
		h.respondWithError(w, err.Error(), commentServiceErrorStatus(err))
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"comments": comments,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.respondWithError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		h.respondWithError(w, "invalid comment id", http.StatusBadRequest)
		return
	}

	var req model.CommentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		h.respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}

	comment, err := h.commentService.UpdateComment(r.Context(), id, &req, userID)
	if err != nil {
		h.respondWithError(w, err.Error(), commentServiceErrorStatus(err))
		return
	}

	h.respondWithJSON(w, http.StatusOK, comment)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.respondWithError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		h.respondWithError(w, "invalid comment id", http.StatusBadRequest)
		return
	}

	if err := h.commentService.DeleteComment(r.Context(), id, userID); err != nil {
		h.respondWithError(w, err.Error(), commentServiceErrorStatus(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CommentHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *CommentHandler) respondWithError(w http.ResponseWriter, message string, code int) {
	type ErrorResponse struct {
		Error string `json:"error"`
	}
	h.respondWithJSON(w, code, ErrorResponse{Error: message})
}
