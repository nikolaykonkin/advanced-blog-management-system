package service

// ActionLogger определяет интерфейс отложенного логгера действий пользователя
// Реализация - pkg/logger.ActionLogger
type ActionLogger interface {
	// Log отправляет событие на запись
	Log(event string)
}
