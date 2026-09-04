package handler

import (
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/service"
	"encoding/json"
	"net/http"
)

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать обработчик создания комментария
	// 1. Получить userID из контекста
	// 2. Если пользователь не авторизован - вернуть 401 Unauthorized
	// 3. Получить postID из URL параметра
	// 4. Распарсить JSON в CommentCreateRequest
	// 5. Валидировать данные (req.Validate())
	// 6. Если валидация не прошла - вернуть 400 Bad Request
	// 7. Вызвать commentService.CreateComment()
	// 8. Если пост не найден - вернуть 404 Not Found
	// 9. Если ошибка - вернуть 500 Internal Server Error
	// 10. Если успешно - вернуть 201 Created с CommentResponse

	var req model.CommentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func (h *CommentHandler) GetCommentsByPostID(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать обработчик получения комментариев к посту
	// 1. Получить postID из URL параметра
	// 2. Получить limit и offset из query параметров
	// 3. Установить defaults: limit=10, offset=0
	// 4. Валидировать параметры
	// 5. Вызвать commentService.GetCommentsByPostID()
	// 6. Получить количество комментариев
	// 7. Если успешно - вернуть 200 OK с массивом комментариев и метаинформацией
	//    Формат: {"comments": [...], "total": 15, "limit": 10, "offset": 0}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать обработчик обновления комментария
	// 1. Получить userID из контекста
	// 2. Если пользователь не авторизован - вернуть 401 Unauthorized
	// 3. Получить commentID из URL параметра
	// 4. Распарсить JSON в CommentUpdateRequest
	// 5. Валидировать данные (req.Validate())
	// 6. Вызвать commentService.UpdateComment()
	// 7. Если комментарий не найден - вернуть 404 Not Found
	// 8. Если пользователь не автор - вернуть 403 Forbidden
	// 9. Если ошибка - вернуть 500 Internal Server Error
	// 10. Если успешно - вернуть 200 OK с обновленным CommentResponse

	var req model.CommentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать обработчик удаления комментария
	// 1. Получить userID из контекста
	// 2. Если пользователь не авторизован - вернуть 401 Unauthorized
	// 3. Получить commentID из URL параметра
	// 4. Вызвать commentService.DeleteComment()
	// 5. Если комментарий не найден - вернуть 404 Not Found
	// 6. Если пользователь не автор - вернуть 403 Forbidden
	// 7. Если ошибка - вернуть 500 Internal Server Error
	// 8. Если успешно - вернуть 204 No Content

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
