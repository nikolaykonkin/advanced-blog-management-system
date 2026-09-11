package middleware

import (
	"advanced-blog-management-system/pkg/auth"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-secret-for-middleware-tests"

// ---------------------------------------------------------------------
// ExtractToken
// ---------------------------------------------------------------------

func TestExtractToken_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer abc.def.ghi")

	token, err := ExtractToken(req)

	require.NoError(t, err)
	assert.Equal(t, "abc.def.ghi", token)
}

// схема сравнивается через strings.EqualFold, регистр не важен
func TestExtractToken_CaseInsensitiveScheme(t *testing.T) {
	for _, scheme := range []string{"bearer", "BEARER", "BeArEr"} {
		t.Run(scheme, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", scheme+" abc.def.ghi")

			token, err := ExtractToken(req)

			require.NoError(t, err)
			assert.Equal(t, "abc.def.ghi", token)
		})
	}
}

func TestExtractToken_TrimsExtraWhitespace(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer   abc.def.ghi  ")

	token, err := ExtractToken(req)

	require.NoError(t, err)
	assert.Equal(t, "abc.def.ghi", token)
}

func TestExtractToken_MissingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := ExtractToken(req)

	require.Error(t, err)
	assert.Equal(t, errMissingAuthHeader, err)
}

func TestExtractToken_MalformedHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"wrong scheme", "Basic dXNlcjpwYXNz"},
		{"no token after Bearer", "Bearer"},
		{"only whitespace as token", "Bearer    "},
		{"no space at all", "Beareraaa.bbb.ccc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tt.header)

			_, err := ExtractToken(req)

			require.Error(t, err)
			assert.Equal(t, errMalformedAuthHeader, err)
		})
	}
}

// ---------------------------------------------------------------------
// AuthMiddleware
// ---------------------------------------------------------------------

func newDownstreamRecorder() (handler http.Handler, called *bool, gotUserID *int) {
	called = new(bool)
	gotUserID = new(int)
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		if id, ok := GetUserIDFromContext(r); ok {
			*gotUserID = id
		}
		w.WriteHeader(http.StatusOK)
	})
	return
}

func TestAuthMiddleware_ValidToken_SetsUserIDAndCallsNext(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	token, _, err := auth.GenerateToken(42, "user@example.com", "user", testJWTSecret)
	require.NoError(t, err)

	next, called, gotUserID := newDownstreamRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(rec, req)

	assert.True(t, *called)
	assert.Equal(t, 42, *gotUserID)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ответ с ошибкой - JSON, как и на уровне хендлеров
func TestAuthMiddleware_MissingHeader_Returns401JSON(t *testing.T) {
	next, called, _ := newDownstreamRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(rec, req)

	assert.False(t, *called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, errMissingAuthHeader.Error(), body["error"])
}

// просроченный токен и мусорная строка должны возвращать одно и то же
// сообщение - точная причина отказа клиенту не раскрывается
func TestAuthMiddleware_InvalidToken_ReturnsGenericMessage(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)

	garbageToken := "this-is-not-a-jwt-at-all"
	wrongSecretToken, _, err := auth.GenerateToken(1, "a@b.com", "a", "a-completely-different-secret")
	require.NoError(t, err)

	for _, tt := range []struct {
		name  string
		token string
	}{
		{"malformed token", garbageToken},
		{"token signed with wrong secret", wrongSecretToken},
	} {
		t.Run(tt.name, func(t *testing.T) {
			next, called, _ := newDownstreamRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rec := httptest.NewRecorder()

			AuthMiddleware(next).ServeHTTP(rec, req)

			assert.False(t, *called)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)

			var body map[string]string
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, "invalid or expired token", body["error"])
		})
	}
}

// ---------------------------------------------------------------------
// OptionalAuthMiddleware
// ---------------------------------------------------------------------

func TestOptionalAuthMiddleware_NoToken_PassesThroughWithoutUserID(t *testing.T) {
	next, called, gotUserID := newDownstreamRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	OptionalAuthMiddleware(next).ServeHTTP(rec, req)

	assert.True(t, *called)
	assert.Equal(t, 0, *gotUserID)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOptionalAuthMiddleware_ValidToken_SetsUserID(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	token, _, err := auth.GenerateToken(7, "user@example.com", "user", testJWTSecret)
	require.NoError(t, err)

	next, called, gotUserID := newDownstreamRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	OptionalAuthMiddleware(next).ServeHTTP(rec, req)

	assert.True(t, *called)
	assert.Equal(t, 7, *gotUserID)
}

// в отличие от AuthMiddleware, невалидный токен здесь не блокирует запрос
func TestOptionalAuthMiddleware_InvalidToken_PassesThroughWithoutUserID(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)

	next, called, gotUserID := newDownstreamRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer garbage-token")
	rec := httptest.NewRecorder()

	OptionalAuthMiddleware(next).ServeHTTP(rec, req)

	assert.True(t, *called)
	assert.Equal(t, 0, *gotUserID)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ---------------------------------------------------------------------
// GetUserIDFromContext
// ---------------------------------------------------------------------

func TestGetUserIDFromContext_NotSet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	id, ok := GetUserIDFromContext(req)

	assert.False(t, ok)
	assert.Equal(t, 0, id)
}

// значение типа not-an-int под UserKey не должно ронять объявление типа
func TestGetUserIDFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserKey, "not-an-int")
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	id, ok := GetUserIDFromContext(req)

	assert.False(t, ok)
	assert.Equal(t, 0, id)
}
