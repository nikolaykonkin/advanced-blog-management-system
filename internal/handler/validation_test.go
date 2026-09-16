package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validationErrorFor прогоняет запрос через Validate() и возвращает уже переведенное сообщение —
// так тест проверяет ровно ту связку, которая работает в хендлерах,
// а не собранную вручную validator.ValidationErrors
func validationErrorFor(t *testing.T, req interface{ Validate() error }) string {
	t.Helper()
	err := req.Validate()
	require.Error(t, err, "ожидалась ошибка валидации")
	return formatValidationError(err)
}

func TestFormatValidationError_RequiredField(t *testing.T) {
	msg := validationErrorFor(t, &model.UserCreateRequest{
		Email:    "john@example.com",
		Password: "password123",
	})

	assert.Contains(t, msg, "username is required")
}

func TestFormatValidationError_InvalidEmail(t *testing.T) {
	msg := validationErrorFor(t, &model.UserCreateRequest{
		Username: "johndoe",
		Email:    "not-an-email",
		Password: "password123",
	})

	assert.Equal(t, "email must be a valid email address", msg)
}

func TestFormatValidationError_MinLength(t *testing.T) {
	msg := validationErrorFor(t, &model.UserCreateRequest{
		Username: "johndoe",
		Email:    "john@example.com",
		Password: "123",
	})

	assert.Equal(t, "password must be at least 6 characters long", msg)
}

func TestFormatValidationError_MaxLength(t *testing.T) {
	msg := validationErrorFor(t, &model.PostCreateRequest{
		Title:   strings.Repeat("a", 201),
		Content: "content",
	})

	assert.Equal(t, "title must be at most 200 characters long", msg)
}

func TestFormatValidationError_OneOf(t *testing.T) {
	msg := validationErrorFor(t, &model.PostUpdateRequest{
		Title:   "title",
		Content: "content",
		Status:  "garbage",
	})

	assert.Equal(t, "status must be one of: draft, published", msg)
}

// несколько непройденных правил сразу склеиваются в одну строку через "; "
func TestFormatValidationError_MultipleErrors(t *testing.T) {
	msg := validationErrorFor(t, &model.UserCreateRequest{})

	assert.Contains(t, msg, "username is required")
	assert.Contains(t, msg, "email is required")
	assert.Contains(t, msg, "password is required")
	assert.Equal(t, 2, strings.Count(msg, "; "), "три сообщения должны быть склеены двумя разделителями")
}

// ошибка не от валидатора не должна утечь к клиенту как есть
func TestFormatValidationError_NonValidatorError(t *testing.T) {
	assert.Equal(t, "invalid request data", formatValidationError(errors.New("some internal failure")))
}

// проверяет, что в ответе API нет имен структур и тегов валидатора
func TestAuthHandler_Register_ValidationError_ReturnsCleanMessage(t *testing.T) {
	h := NewAuthHandler(service.NewUserService(newFakeUserRepo()), "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(
		`{"username":"johndoe","email":"not-an-email","password":"password123"}`))
	rec := httptest.NewRecorder()

	h.RegisterHandler(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var resp struct {
		Error string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	assert.Equal(t, "email must be a valid email address", resp.Error)
	assert.NotContains(t, resp.Error, "UserCreateRequest")
	assert.NotContains(t, resp.Error, "tag")
	assert.NotContains(t, resp.Error, "Key:")
}
