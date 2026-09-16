package handler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// formatValidationError переводит ошибку валидатора в короткое сообщение для клиента:
// стандартный err.Error() раскрывает внутренние имена структур и тегов,
// а на не-валидаторную ошибку отдаем общее сообщение, ничего не раскрывая
func formatValidationError(err error) string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return "invalid request data"
	}

	messages := make([]string, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		messages = append(messages, fieldErrorMessage(fieldErr))
	}
	return strings.Join(messages, "; ")
}

// fieldErrorMessage формирует сообщение по одному непройденному правилу
// Пока все валидируемые поля однословные, strings.ToLower(Field()) совпадает с json-тегом
func fieldErrorMessage(fieldErr validator.FieldError) string {
	field := strings.ToLower(fieldErr.Field())

	switch fieldErr.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		// все поля с min/max в проекте — строки, поэтому единица измерения всегда символы
		// для числового поля здесь понадобился бы отдельный текст по fieldErr.Kind()
		return fmt.Sprintf("%s must be at least %s characters long", field, fieldErr.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, fieldErr.Param())
	case "oneof":
		allowed := strings.ReplaceAll(fieldErr.Param(), " ", ", ")
		return fmt.Sprintf("%s must be one of: %s", field, allowed)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
