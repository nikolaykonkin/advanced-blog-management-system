package apperrors

import (
	"errors"
	"net/http"
)

var (
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user with this email or username already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrPostNotFound       = errors.New("post not found")
	ErrCommentNotFound    = errors.New("comment not found")
	ErrInvalidPostID      = errors.New("invalid post id")
)

// ToHTTPStatus сопоставляет ошибку приложения с HTTP статус-кодом
// Неизвестные ошибки сопоставляются с 500
func ToHTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrPostNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrCommentNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrUserAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusBadRequest
	case errors.Is(err, ErrInvalidPostID):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
