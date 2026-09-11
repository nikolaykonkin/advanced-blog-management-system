package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------
// RecoveryMiddleware
// ---------------------------------------------------------------------

// assert.NotPanics реально проверяет recover(), а не только ответ
func TestRecoveryMiddleware_RecoversFromPanic(t *testing.T) {
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went terribly wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		RecoveryMiddleware(panicking).ServeHTTP(rec, req)
	})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Internal Server Error")
}

func TestRecoveryMiddleware_NoPanic_PassesThroughUnchanged(t *testing.T) {
	normal := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("all good"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	RecoveryMiddleware(normal).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "all good", rec.Body.String())
}

// ---------------------------------------------------------------------
// CORSMiddleware
// ---------------------------------------------------------------------

func TestCORSMiddleware_SetsHeadersOnNormalRequest(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	CORSMiddleware(next).ServeHTTP(rec, req)

	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization", rec.Header().Get("Access-Control-Allow-Headers"))
}

// проверяется через флаг called, не только через код ответа - 200
// мог бы прийти и от самого next
func TestCORSMiddleware_OPTIONS_ShortCircuitsWithoutCallingNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec := httptest.NewRecorder()

	CORSMiddleware(next).ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_NonOPTIONS_CallsNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	CORSMiddleware(next).ServeHTTP(rec, req)

	assert.True(t, called)
}

// ---------------------------------------------------------------------
// LoggingMiddleware
// ---------------------------------------------------------------------

func TestLoggingMiddleware_PreservesStatusCodeAndCallsNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})

	req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
	rec := httptest.NewRecorder()

	LoggingMiddleware(next).ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusTeapot, rec.Code)
}

func TestLoggingMiddleware_DefaultsToOKWhenHandlerNeverCallsWriteHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("no explicit WriteHeader call"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	LoggingMiddleware(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
