package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken создает новый JWT токен
func GenerateToken(userID int, email, username, secret string) (string, time.Time, error) {
	// TODO: Реализовать генерацию JWT токена
	// 1. Установить время истечения токена (например, 24 часа)
	// 2. Создать объект Claims с user_id, email, username и сроком действия
	// 3. Создать новый JWT токен с Claims
	// 4. Подписать токен используя HS256 и secret
	// 5. Вернуть строку токена и время истечения
	return "", time.Time{}, nil
}

// ValidateToken проверяет и распарсивает JWT токен
func ValidateToken(tokenString, secret string) (*Claims, error) {
	// TODO: Реализовать валидацию JWT токена
	// 1. Распарсить токен используя jwt.ParseWithClaims
	// 2. Если ошибка парсинга - вернуть ошибку
	// 3. Проверить что токен валидный (claims.Valid())
	// 4. Если невалидный - вернуть ошибку
	// 5. Вернуть Claims из токена
	return nil, errors.New("token validation not implemented")
}
