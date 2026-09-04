package middleware

import (
	"net/http"
)

// UserContextKey используется для сохранения данных пользователя в context
type UserContextKey string

const UserKey UserContextKey = "user"

// AuthMiddleware проверяет JWT токен и добавляет данные пользователя в context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Реализовать middleware аутентификации
		// 1. Получить токен из заголовка Authorization (Bearer <token>)
		// 2. Если токена нет или формат неверный - вернуть 401
		// 3. Валидировать токен (использовать jwt.ParseWithClaims)
		// 4. Если токен невалидный - вернуть 401
		// 5. Извлечь user_id из claims
		// 6. Добавить user_id в context запроса
		// 7. Передать request дальше в цепь middleware
		next.ServeHTTP(w, r)
	})
}

// OptionalAuthMiddleware проверяет JWT токен если он присутствует, но не обязателен
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Реализовать опциональное middleware аутентификации
		// 1. Получить токен из заголовка Authorization если присутствует
		// 2. Если токен есть - валидировать его
		// 3. Если токен валидный - добавить user_id в context
		// 4. Если токена нет или он невалидный - просто передать request дальше
		// 5. Передать request дальше в цепь middleware
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext извлекает пользователя из context
func GetUserIDFromContext(r *http.Request) (int, bool) {
	// TODO: Реализовать извлечение user_id из context
	// Получить значение из context по ключу UserKey
	// Попробовать привести его к int
	// Вернуть user_id и флаг успеха
	return 0, false
}

// ExtractToken извлекает JWT токен из заголовка Authorization
func ExtractToken(r *http.Request) (string, error) {
	// TODO: Реализовать извлечение токена из заголовка
	// 1. Получить заголовок Authorization
	// 2. Проверить что он начинается с "Bearer "
	// 3. Извлечь токен после "Bearer "
	// 4. Вернуть токен или ошибку если формат неверный
	return "", nil
}
