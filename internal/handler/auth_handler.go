package handler

import (
	"advanced-blog-management-system/internal/service"
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
	// TODO: Реализовать обработчик регистрации
	// 1. Распарсить JSON из тела запроса в структуру UserCreateRequest
	//    Используйте json.NewDecoder(r.Body).Decode(&req)
	// 2. Обработайте ошибку парсинга - вернуть 400 Bad Request
	// 3. Валидировать входные данные (req.Validate())
	// 4. Вызвать userService.Register(r.Context(), &req)
	// 5. Обработайте различные ошибки:
	//    - Если пользователь уже существует → 400 или 409
	//    - Если другая ошибка сервера → 500
	// 6. Если успешно:
	//    - Сгенерируйте JWT токен используя pkg/auth.GenerateToken()
	//    - Создайте TokenResponse со статусом
	//    - Вернуть 201 Created с ответом в JSON
	//
	// Статус коды:
	// - 201 Created - пользователь успешно создан
	// - 400 Bad Request - некорректные данные или пользователь уже существует
	// - 500 Internal Server Error - ошибка сервера
}

// LoginHandler обрабатывает POST /api/login
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать обработчик входа
	// 1. Распарсить JSON из тела запроса в структуру UserLoginRequest
	// 2. Обработайте ошибку парсинга - вернуть 400 Bad Request
	// 3. Валидировать входные данные (req.Validate())
	// 4. Вызвать userService.Login(r.Context(), &req)
	// 5. Обработайте ошибки:
	//    - Если пользователь не найден или пароль неверный → 401 Unauthorized
	//    - Если другая ошибка → 500
	// 6. Если успешно:
	//    - Сгенерируйте JWT токен используя pkg/auth.GenerateToken()
	//    - Создайте TokenResponse со статусом
	//    - Вернуть 200 OK с ответом в JSON
	//
	// Статус коды:
	// - 200 OK - успешный вход
	// - 400 Bad Request - некорректные данные
	// - 401 Unauthorized - неверные учетные данные
	// - 500 Internal Server Error - ошибка сервера
}

// respondWithJSON - helper для отправки JSON ответов
func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// TODO: Реализовать отправку JSON ответа
	// 1. Установите заголовок Content-Type: application/json
	// 2. Установите HTTP статус код
	// 3. Закодируйте payload в JSON используя json.NewEncoder(w).Encode(payload)
}

// respondWithError - helper для отправки ошибок
func (h *AuthHandler) respondWithError(w http.ResponseWriter, message string, code int) {
	// TODO: Реализовать отправку ошибки
	// 1. Создайте структуру с полем Error
	// 2. Используйте respondWithJSON для отправки
	type ErrorResponse struct {
		Error string `json:"error"`
	}
	_ = ErrorResponse{Error: message}
}
