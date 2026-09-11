package apperrors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestToHTTPStatus_KnownErrors защищает от того, что при добавлении новой ошибки в пакет
// забудут добавить case в ToHTTPStatus — тогда она молча уйдёт в default (500)
func TestToHTTPStatus_KnownErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"ErrUnauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"ErrForbidden", ErrForbidden, http.StatusForbidden},
		{"ErrUserNotFound", ErrUserNotFound, http.StatusNotFound},
		{"ErrPostNotFound", ErrPostNotFound, http.StatusNotFound},
		{"ErrCommentNotFound", ErrCommentNotFound, http.StatusNotFound},
		{"ErrUserAlreadyExists", ErrUserAlreadyExists, http.StatusConflict},
		// 401, не 400: запрос синтаксически корректен, но не прошел аутентификацию (RFC 9110)
		{"ErrInvalidCredentials_mapsTo401NotBadRequest", ErrInvalidCredentials, http.StatusUnauthorized},
		{"ErrInvalidPostID", ErrInvalidPostID, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantStatus, ToHTTPStatus(tt.err))
		})
	}
}

// default-ветка: ошибка вне известного списка маппится в 500
func TestToHTTPStatus_UnknownError_MapsToInternalServerError(t *testing.T) {
	assert.Equal(t, http.StatusInternalServerError, ToHTTPStatus(errors.New("something exploded")))
}

// В коде ToHTTPStatus никогда не вызывается с nil, но фиксируем контракт:
// errors.Is(nil, X) = false, поэтому nil уходит в default, а не паникует
func TestToHTTPStatus_NilError_MapsToInternalServerError(t *testing.T) {
	assert.Equal(t, http.StatusInternalServerError, ToHTTPStatus(nil))
}

// Сервисы оборачивают известные ошибки через fmt.Errorf("%w", err) —
// errors.Is обязан узнавать их сквозь обертку, иначе хендлер отдаст 500 вместо нужного статуса
func TestToHTTPStatus_WrappedError_StillMatches(t *testing.T) {
	wrapped := fmt.Errorf("failed to get post: %w", ErrPostNotFound)

	assert.Equal(t, http.StatusNotFound, ToHTTPStatus(wrapped))
}

// TestToHTTPStatus_DoublyWrappedError_StillMatches — то же самое, но через две
// обертки подряд, как реально бывает в UserService.Register
func TestToHTTPStatus_DoublyWrappedError_StillMatches(t *testing.T) {
	wrapped := fmt.Errorf("registration failed: %w", fmt.Errorf("create user: %w", ErrUserAlreadyExists))

	assert.Equal(t, http.StatusConflict, ToHTTPStatus(wrapped))
}
