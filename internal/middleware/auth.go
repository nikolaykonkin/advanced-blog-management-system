package middleware

import (
	"advanced-blog-management-system/pkg/auth"
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
)

// UserContextKey — типизированный ключ для контекста, защищает от коллизий
type UserContextKey string

// UserKey — ключ для хранения user_id в контексте запроса
const UserKey UserContextKey = "user"

var (
	errMissingAuthHeader   = errors.New("authorization header is required")
	errMalformedAuthHeader = errors.New("authorization header must be in the format: Bearer <token>")
)

// jwtSecret читает секрет для проверки JWT из переменной окружения
// Сигнатуры AuthMiddleware/OptionalAuthMiddleware заданы шаблоном без параметра секрета,
// поэтому он читается напрямую здесь, а не передаётся через конструктор
func jwtSecret() string {
	return os.Getenv("JWT_SECRET")
}

// AuthMiddleware проверяет JWT токен и добавляет данные пользователя в context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := ExtractToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(token, jwtSecret())
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware проверяет JWT токен если он присутствует, но не обязателен
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := ExtractToken(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := auth.ValidateToken(token, jwtSecret())
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext извлекает user_id из context
func GetUserIDFromContext(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(UserKey).(int)
	return id, ok
}

// ExtractToken извлекает JWT токен из заголовка Authorization
func ExtractToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", errMissingAuthHeader
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", errMalformedAuthHeader
	}

	return strings.TrimSpace(parts[1]), nil
}
