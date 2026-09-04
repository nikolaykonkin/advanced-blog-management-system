package middleware

import (
	"net/http"
)

// LoggingMiddleware логирует информацию о каждом запросе
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Реализовать логирование запросов
		// 1. Получить текущее время (time.Now())
		// 2. Обернуть ResponseWriter чтобы перехватить статус код
		//    (создайте структуру которая реализует http.ResponseWriter интерфейс)
		// 3. Передать request дальше (next.ServeHTTP)
		// 4. После выполнения - залогировать:
		//    - IP адрес клиента (r.RemoteAddr)
		//    - HTTP метод и путь (r.Method, r.RequestURI)
		//    - Статус код ответа (из обернутого ResponseWriter)
		//    - Время выполнения (time.Since(start))
		// 5. Пример лога: "127.0.0.1 | GET /api/posts | 200 | 12.5ms"

		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware восстанавливает приложение после паники
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Реализовать восстановление после паники
		// 1. Использовать defer и recover() для перехвата паники
		// 2. Если паника произошла:
		//    - Залогировать ошибку используя log.Printf
		//    - Вернуть HTTP 500 Internal Server Error
		//    - Использовать http.Error() для отправки ошибки
		// 3. Если паники нет - передать request дальше (next.ServeHTTP)
		// 4. Пример структуры:
		//    defer func() {
		//        if err := recover(); err != nil {
		//            // логирование и ответ
		//        }
		//    }()

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware добавляет CORS заголовки к ответу
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Реализовать CORS
		// 1. Добавить заголовки ответа:
		//    - Access-Control-Allow-Origin: * (или конкретный домен)
		//    - Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
		//    - Access-Control-Allow-Headers: Content-Type, Authorization
		// 2. Если HTTP метод OPTIONS - вернуть 200 OK и остановить обработку
		// 3. Иначе передать request дальше (next.ServeHTTP)
		// 4. Используйте w.Header().Set() для установки заголовков

		next.ServeHTTP(w, r)
	})
}

// responseWriter обертка для ResponseWriter чтобы перехватить статус код
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
