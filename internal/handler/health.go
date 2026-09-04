package handler

import (
	"net/http"
)

// HealthCheckHandler обрабатывает GET /api/health
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать проверку здоровья приложения
	// 1. Установить заголовок Content-Type: application/json
	// 2. Написать HTTP статус код 200 OK
	// 3. Вернуть JSON ответ с полем "status": "ok"
	// Пример ответа: {"status":"ok"}
}
