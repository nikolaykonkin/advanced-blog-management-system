package main

import (
	"log"
)

func main() {
	// TODO: Загрузить переменные окружения из .env файла
	// Используйте github.com/joho/godotenv для загрузки .env

	// TODO: Прочитать конфигурацию из переменных окружения
	// Ожидаемые переменные окружения:
	// - DB_HOST (по умолчанию: localhost)
	// - DB_PORT (по умолчанию: 5432)
	// - DB_USER (по умолчанию: postgres)
	// - DB_PASSWORD (по умолчанию: postgres)
	// - DB_NAME (по умолчанию: blog_db)
	// - DB_SSLMODE (по умолчанию: disable)
	// - JWT_SECRET (ОБЯЗАТЕЛЬНАЯ - секретный ключ для JWT)
	// - SERVER_HOST (по умолчанию: 0.0.0.0)
	// - SERVER_PORT (по умолчанию: 8080)
	//
	// Подсказка: используйте os.Getenv() с fallback значениями
	// или strconv для преобразования портов в числа

	// TODO: Создать подключение к PostgreSQL
	// Используйте функцию database.NewPostgresDB из pkg/database
	// Передайте конфигурацию БД
	// Обработайте возможные ошибки подключения

	// TODO: Выполнить миграции БД
	// Прочитайте SQL файлы из папки migrations/
	// Выполните их в нужном порядке:
	// 1. 001_init_schema.sql (создание таблиц)
	// 2. 002_add_indexes.sql (создание индексов)
	// Используйте функцию database.RunMigrations()

	// TODO: Инициализировать репозитории
	// Создайте экземпляры:
	// - UserRepository
	// - PostRepository
	// - CommentRepository
	// Передайте им подключение к БД

	// TODO: Инициализировать сервисы
	// Создайте экземпляры:
	// - UserService (передайте userRepository)
	// - PostService (передайте postRepository, userRepository, commentRepository)
	// - CommentService (передайте commentRepository, postRepository, userRepository)

	// TODO: Инициализировать обработчики (handlers)
	// Создайте экземпляры:
	// - AuthHandler (передайте userService и JWT_SECRET)
	// - PostHandler (передайте postService)
	// - CommentHandler (передайте commentService)

	// TODO: Создать и настроить HTTP роутер
	// Вызовите функцию setupRouter() которая вернет chi.Mux
	// Роутер должен содержать:
	// - Все middleware (логирование, recovery, CORS, аутентификация где нужна)
	// - Все HTTP эндпоинты согласно спецификации

	// TODO: Создать HTTP сервер
	// Создайте структуру http.Server с:
	// - Addr: полученный из конфигурации адрес и порт
	// - Handler: роутер
	// - ReadTimeout: 15 секунд
	// - WriteTimeout: 15 секунд
	// - IdleTimeout: 60 секунд

	// TODO: Запустить сервер в отдельной горутине
	// go func() { ... }()
	// Обработайте ошибку http.ErrServerClosed как успех (это нормально при shutdown)

	// TODO: Установить обработчик сигналов завершения
	// Перехватите сигналы SIGINT и SIGTERM (ctrl+C, kill и т.д.)
	// Используйте os.Signal и signal.Notify()

	// TODO: Graceful shutdown
	// Когда получен сигнал завершения:
	// 1. Логируйте информацию о завершении
	// 2. Создайте контекст с таймаутом (5-10 секунд)
	// 3. Вызовите srv.Shutdown(ctx) для корректного завершения
	// 4. Закройте подключение к БД: db.Close()
	// 5. Выйдите из программы

	// TODO: Логирование
	// В главной функции логируйте ключевые события:
	// - Загрузка конфигурации
	// - Подключение к БД
	// - Выполнение миграций
	// - Запуск сервера (какой адрес и порт)
	// - Получение сигнала завершения
	// - Результат shutdown

	log.Println("Server starting... (TODO: implement main.go)")
}

func setupRouter() interface{} {
	// TODO: Создать chi роутер
	// Используйте chi.NewRouter()

	// TODO: Зарегистрировать middleware в порядке:
	// 1. LoggingMiddleware (должен быть первым)
	// 2. RecoveryMiddleware
	// 3. CORSMiddleware

	// TODO: Зарегистрировать все публичные эндпоинты:
	// - GET /api/health (HealthCheckHandler)
	// - POST /api/register
	// - POST /api/login
	// - GET /api/posts
	// - GET /api/posts/{id}
	// - GET /api/users/{id}/posts
	// - GET /api/posts/{postId}/comments

	// TODO: Зарегистрировать защищенные эндпоинты (требуют AuthMiddleware):
	// - POST /api/posts
	// - PUT /api/posts/{id}
	// - DELETE /api/posts/{id}
	// - POST /api/posts/{postId}/comments
	// - PUT /api/comments/{id}
	// - DELETE /api/comments/{id}
	//
	// Подсказка для защиты: используйте r.With(middleware.AuthMiddleware)
	// или создайте отдельный роут-группу для защищенных эндпоинтов

	// TODO: Вернуть настроенный роутер
	return nil
}
