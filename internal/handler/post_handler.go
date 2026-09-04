package handler

import (
	"advanced-blog-management-system/internal/service"
	"encoding/json"
	"net/http"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать обработчик создания поста
	// 1. Получить userID из контекста (middleware должно его добавить)
	// 2. Распарсить JSON в PostCreateRequest
	// 3. Валидировать данные (req.Validate())
	// 4. Если валидация не прошла - вернуть 400 Bad Request
	// 5. Вызвать postService.CreatePost()
	// 6. Если ошибка - вернуть 500 Internal Server Error
	// 7. Если успешно - вернуть 201 Created с PostResponse

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать обработчик получения поста по ID
	// 1. Извлечь ID поста из URL параметра (используя chi.URLParam или похожее)
	// 2. Конвертировать ID в число
	// 3. Вызвать postService.GetPost()
	// 4. Если пост не найден - вернуть 404 Not Found
	// 5. Если ошибка - вернуть 500 Internal Server Error
	// 6. Если успешно - вернуть 200 OK с PostResponse

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать обработчик получения всех постов
	// 1. Получить параметры limit и offset из query string
	// 2. Установить defaults: limit=10, offset=0
	// 3. Валидировать параметры (limit должен быть > 0 и <= 100)
	// 4. Вызвать postService.GetAllPosts()
	// 5. Получить количество постов через postService.GetPostsCount()
	// 6. Вернуть 200 OK с объектом:
	//    {
	//      "posts": [...],
	//      "total": 42,
	//      "limit": 10,
	//      "offset": 0
	//    }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать обработчик обновления поста
	// 1. Получить userID из контекста
	// 2. Извлечь ID поста из URL параметра
	// 3. Распарсить JSON в PostUpdateRequest
	// 4. Валидировать данные
	// 5. Вызвать postService.UpdatePost()
	// 6. Если пост не найден - вернуть 404 Not Found
	// 7. Если пользователь не автор - вернуть 403 Forbidden
	// 8. Если ошибка - вернуть 500 Internal Server Error
	// 9. Если успешно - вернуть 200 OK с обновленным PostResponse

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать обработчик удаления поста
	// 1. Получить userID из контекста
	// 2. Извлечь ID поста из URL параметра
	// 3. Вызвать postService.DeletePost()
	// 4. Если пост не найден - вернуть 404 Not Found
	// 5. Если пользователь не автор - вернуть 403 Forbidden
	// 6. Если ошибка - вернуть 500 Internal Server Error
	// 7. Если успешно - вернуть 204 No Content

	w.WriteHeader(http.StatusNoContent)
}

func (h *PostHandler) GetPostsByAuthor(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать обработчик получения постов автора
	// 1. Извлечь ID автора из URL параметра
	// 2. Получить параметры limit и offset из query string
	// 3. Установить defaults: limit=10, offset=0
	// 4. Вызвать postService.GetPostsByAuthor()
	// 5. Вернуть 200 OK с объектом постов и метаинформацией

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
