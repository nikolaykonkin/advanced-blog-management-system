# Advanced Blog Management System

![CI](https://github.com/nikolaykonkin/advanced-blog-management-system/actions/workflows/ci.yml/badge.svg)

REST API для блог-платформы на Go: регистрация и вход пользователей, посты (включая отложенную публикацию), комментарии, JWT-аутентификация и отложенное логирование действий пользователя через канал и фоновую горутину.

Дипломный проект по программе «Go-разработчик с нуля» (Нетология), выполнен на основе шаблона `gopr-temp-ex-main`.

## Содержание

- [Архитектура](#архитектура)
- [Технологический стек](#технологический-стек)
- [Быстрый старт](#быстрый-старт)
- [Переменные окружения](#переменные-окружения)
- [API эндпоинты](#api-эндпоинты)
- [Пример сценария через curl](#пример-сценария-через-curl)
- [Конкурентность](#конкурентность)
- [Решения и их обоснование](#решения-и-их-обоснование)
- [Тестирование](#тестирование)
- [Docker](#docker)
- [Известные ограничения](#известные-ограничения)
- [Частые проблемы](#частые-проблемы)
- [Возможные доработки](#возможные-доработки)

## Архитектура

Слоистая архитектура с однонаправленной зависимостью сверху вниз:

```
HTTP-запрос
    │
    ▼
Middleware (Logging, Recovery, CORS, Auth)
    │
    ▼
Handler (internal/handler)       — парсинг JSON, HTTP-коды, JSON-ответы
    │
    ▼
Service (internal/service)       — бизнес-логика, валидация, права доступа
    │
    ▼
Repository (internal/repository) — SQL-запросы к PostgreSQL
```

`Service` зависит от интерфейсов репозиториев (`internal/repository/interfaces.go`), а не от конкретной реализации на `database/sql` — это и позволяет подменять их фейками в юнит-тестах без поднятия настоящей БД.

Структура директорий:

```
advanced-blog-management-system/
├── cmd/
│   └── api/
│       └── main.go                 # точка входа: конфиг, DI, роутинг, graceful shutdown
├── internal/
│   ├── errors/
│   │   └── apperrors/
│   │       └── errors.go           # sentinel-ошибки и маппинг в HTTP-статусы
│   ├── handler/                    # HTTP-обработчики
│   │   ├── auth_handler.go
│   │   ├── post_handler.go
│   │   ├── comment_handler.go
│   │   └── health.go
│   ├── middleware/                 # middleware уровня HTTP
│   │   ├── auth.go                 # JWT-аутентификация
│   │   └── logging.go              # логирование, recovery, CORS
│   ├── model/                      # модели и DTO
│   │   └── models.go
│   ├── repository/                 # слой доступа к данным
│   │   ├── errors.go               # sentinel-ошибки (not found)
│   │   ├── interfaces.go
│   │   ├── user_repo.go
│   │   ├── post_repo.go
│   │   └── comment_repo.go
│   └── service/                    # бизнес-логика
│       ├── user_service.go
│       ├── post_service.go
│       ├── comment_service.go
│       └── logger.go               # интерфейс ActionLogger
├── pkg/
│   ├── auth/                       # bcrypt и JWT
│   │   ├── jwt.go
│   │   └── password.go
│   ├── database/                   # подключение к PostgreSQL
│   │   └── db.go
│   └── logger/                     # отложенный логгер (канал + горутина)
│       └── action_logger.go
├── migrations/                     # SQL-миграции
│   ├── 001_init_schema.sql
│   └── 002_add_indexes.sql
├── .env.example                    # пример конфигурации
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── README.md
```

## Технологический стек

| Компонент | Библиотека | Назначение |
|---|---|---|
| Роутинг | [go-chi/chi](https://github.com/go-chi/chi) | HTTP-роутер, группировка middleware |
| БД | PostgreSQL 15 + [lib/pq](https://github.com/lib/pq) | хранение данных |
| Аутентификация | [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) | генерация и валидация JWT (HS256) |
| Хеширование паролей | [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | bcrypt |
| Валидация | [go-playground/validator](https://github.com/go-playground/validator) | валидация DTO по тегам структур |
| Конфигурация | [joho/godotenv](https://github.com/joho/godotenv) | загрузка `.env` |
| Тесты | [stretchr/testify](https://github.com/stretchr/testify) | assert/require в юнит-тестах |

Go 1.26, Docker + Docker Compose для контейнеризации.

## Быстрый старт

### Требования

- Go 1.26+
- Docker и Docker Compose

### Через Docker (рекомендуется)

```bash
cp .env.example .env
docker-compose up --build
```

Приложение поднимется на `http://localhost:8080`, миграции применяются автоматически при старте.

### Локально (без Docker)

```bash
cp .env.example .env
docker-compose up -d db     # только БД
go mod download
go run cmd/api/main.go
```

## Переменные окружения

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `DB_HOST` | `localhost` | хост PostgreSQL |
| `DB_PORT` | `5432` | порт PostgreSQL |
| `DB_USER` | `postgres` | пользователь БД |
| `DB_PASSWORD` | `postgres` | пароль БД |
| `DB_NAME` | `blog_db` | имя базы данных |
| `DB_SSLMODE` | `disable` | режим SSL для подключения |
| `JWT_SECRET` | — (обязательна) | секрет для подписи JWT, приложение не стартует без нее |
| `SERVER_HOST` | `0.0.0.0` | адрес, на котором слушает сервер |
| `SERVER_PORT` | `8080` | порт сервера |
| `ACTION_LOG_PATH` | `log.txt` | путь к файлу отложенного лога действий |

## API эндпоинты

### Публичные

| Метод | Путь | Описание |
|---|---|---|
| GET | `/api/health` | проверка работоспособности |
| POST | `/api/register` | регистрация пользователя |
| POST | `/api/login` | вход, выдача JWT |
| GET | `/api/posts` | список постов (пагинация `?limit=&offset=`, по умолчанию 10/0) |
| GET | `/api/posts/{id}` | пост по ID |
| GET | `/api/users/{authorID}/posts` | посты конкретного автора |
| GET | `/api/posts/{id}/comments` | комментарии к посту |

### Защищенные (`Authorization: Bearer <token>`)

| Метод | Путь | Описание |
|---|---|---|
| POST | `/api/posts` | создать пост (сразу опубликован либо черновик, если `publish_at` в будущем) |
| PUT | `/api/posts/{id}` | обновить пост (только автор) |
| DELETE | `/api/posts/{id}` | удалить пост вместе со всеми его комментариями (только автор) |
| POST | `/api/posts/{id}/comments` | добавить комментарий (пост должен быть опубликован) |
| PUT | `/api/comments/{id}` | обновить комментарий (только автор) |
| DELETE | `/api/comments/{id}` | удалить комментарий (только автор) |

Итого 13 эндпоинтов: 7 публичных + 6 защищенных.

## Пример сценария через curl

```bash
# Health check
curl http://localhost:8080/api/health

# Регистрация
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"johndoe","email":"john@example.com","password":"password123"}'

# Вход - в ответе будет token
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'

# Создание поста (замените TOKEN на значение из ответа /login)
curl -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{"title":"My First Post","content":"Hello, world!"}'

# Отложенная публикация - вернется как черновик, планировщик опубликует его
# через 30 сек после наступления publish_at
curl -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{"title":"Scheduled Post","content":"...","publish_at":"2026-01-01T12:00:00Z"}'

# Комментарий к посту (замените POST_ID на ID из ответа выше)
curl -X POST http://localhost:8080/api/posts/POST_ID/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{"content":"Nice post!"}'
```

### Проверка данных через psql

```bash
docker-compose exec db psql -U postgres -d blog_db
```

Внутри `psql`:

```sql
\dt
SELECT id, username, email, created_at FROM users;
SELECT id, title, status, author_id FROM posts;
SELECT id, content, post_id, author_id FROM comments;
```

## Конкурентность

Проект использует два независимых механизма, каждый — отдельная фоновая горутина, запущенная из `main()`.

### Планировщик отложенной публикации

`runScheduler` (`cmd/api/main.go`) каждые 30 секунд вызывает `PostService.PublishScheduledPosts`, которая публикует все черновики, у которых `publish_at` уже наступил. Останавливается через `context.Context`, отменяемый при получении сигнала завершения (`SIGINT`/`SIGTERM`).

### Отложенное логирование действий пользователя

`pkg/logger.ActionLogger` — при создании поста или комментария в канал (`chan string`, буфер 256) отправляется строка-событие; фоновая горутина-воркер вычитывает канал, выдерживает случайную задержку 1-2 секунды и дописывает событие в файл (`log.txt` по умолчанию, путь задается `ACTION_LOG_PATH`). При остановке сервера `Close()` закрывает канал и ждет (`sync.WaitGroup`), пока воркер дозапишет все уже отправленные события, и только после этого закрывает файл.

## Решения и их обоснование

**`ErrInvalidCredentials` → HTTP 401, а не 400.** Согласно RFC 9110, 400 означает синтаксически некорректный запрос; неверный email/пароль — это провал аутентификации при синтаксически корректном запросе, что соответствует 401.

**Один и тот же ответ на "email не найден" и "неверный пароль".** И при логине (`UserService.Login`), и при проверке JWT в `AuthMiddleware` возвращается одинаковое сообщение независимо от точной причины отказа — это защита от user enumeration: иначе по разнице в ответах можно было бы перебором узнать, какие email зарегистрированы в системе.

**Гонка при регистрации.** Между проверкой `ExistsByEmail`/`ExistsByUsername` и вставкой `Create` два параллельных запроса теоретически могут пройти проверку одновременно. UNIQUE-constraint в БД в этом случае вернет ошибку прямо из `Create`, и `UserService.Register` распознает именно `repository.ErrDuplicateUser`, возвращая тот же `ErrUserAlreadyExists`, что и при обычной проверке — с точки зрения клиента это неотличимый, тот же самый 409.

**`author_id` в комментарии не проверяется на существование.** `CommentService.CreateComment` берет `authorID` из подписанного JWT и не делает отдельный запрос в `users`, полагаясь на гарантию подписи. Если пользователь был удален уже после выдачи токена (токен живет 24 часа), `INSERT` упадет на внешнем ключе `comments.author_id → users.id`, и это дойдет до клиента как 500 — редкий и приемлемый для этого проекта случай.

**`ON DELETE CASCADE` вместо явного удаления комментариев в репозитории.** Внешние ключи `posts.author_id`, `comments.post_id`, `comments.author_id` объявлены с `ON DELETE CASCADE` — это подстраховка на уровне БД. При этом `PostService.DeletePost` все равно явно и постранично удаляет комментарии перед удалением поста (см. `deleteAllCommentsForPost`), чтобы не полагаться только на каскад и держать логику удаления на уровне сервиса, а не только в схеме.

**`PublishScheduledPosts` не прерывается на ошибке одного поста.** Ошибки по отдельным постам собираются в срез и объединяются через `errors.Join` в конце — иначе один "битый" пост блокировал бы публикацию остальных, уже готовых, постов на каждом тике планировщика.

**Постраничное удаление комментариев с `offset`, всегда равным 0.** После удаления очередной страницы эти строки исчезают из таблицы, и следующий запрос с тем же `offset=0` видит уже новую "первую страницу" оставшихся комментариев, а не пропускает часть из них, как было бы при обычной постраничной навигации по неизменным данным. Число итераций ограничено (`maxCommentDeletionIterations`) на случай, если `Delete` когда-нибудь сообщит об успехе, ничего не удалив на самом деле.

**Sentinel-ошибки в репозиториях для not-found.** `Update`/`Delete`/`PublishPost` возвращают `repository.ErrXxxNotFound` через `%w` при `RowsAffected() == 0` — это дает HTTP-слою материал для 404, а не 500. Маппинг `repository.ErrXxxNotFound` → `apperrors.ErrXxxNotFound` происходит в сервисах, а не в `apperrors.ToHTTPStatus`: последнее потребовало бы импорта `repository` в `apperrors` и инверсии слоев - `apperrors` не должен знать про персистентность. В сервисах двойная защита: `GetByID` до `Update`/`Delete` ловит обычный случай, а маппинг sentinel-ошибки — защита от race condition, когда запись удалили между `GetByID` и мутацией.

**Пароли хешируются через bcrypt с cost=10.** Cost=10 — стандартное значение по умолчанию, дающее разумный баланс между стойкостью к перебору и задержкой при логине; в production его можно поднять до 12 ценой более медленного логина.

## Тестирование

```bash
go test ./...              # запустить все тесты
go test ./... -v           # с подробным выводом
go test ./... -cover       # с покрытием
go test ./... -race        # проверка на race conditions
```

### Что и как тестируется

Проект покрыт юнит-тестами на уровне сервисов, моделей, middleware и утилит. Вместо mock-библиотек используются in-memory фейки репозиториев — стандартный подход в Go-проектах для проверки реального поведения без внешних зависимостей. В `internal/service` фейки объявлены один раз в `fakes_test.go` и переиспользуются между тестами; в `internal/handler` — свои локальные фейки, потому что хендлеры принимают конкретные типы сервисов, а не интерфейсы.

Покрытие по пакетам (`go test ./... -cover`):

| Пакет | Покрытие |
|---|---|
| `internal/errors/apperrors` | 100% |
| `internal/middleware` | 100% |
| `internal/service` | 96.1% |
| `pkg/logger` | 94.7% |
| `internal/model` | 100% |
| `pkg/auth` | 82.6% |
| `internal/handler` | 30.8% |
| `internal/repository` | 0% (требует PostgreSQL) |
| `pkg/database` | 0% (требует PostgreSQL) |
| `cmd/api` | 0% (точка входа) |

Общее число, взвешенное по количеству строк (а не среднее по пакетам), для этих шести пакетов — **96.0%** (statements), измерено:

```bash
go test ./internal/errors/... ./internal/middleware/... ./internal/service/... \
        ./internal/model/... ./pkg/auth/... ./pkg/logger/... \
        -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -1
```

### Что покрыто и что нет

Покрыто: бизнес-логика сервисов (права доступа, проверки существования, обработка ошибок репозиториев, вызов логгера действий), валидация DTO и бизнес-методы моделей, JWT-аутентификация и остальные middleware, хеширование и токены в `pkg/auth`, отложенная запись в `pkg/logger`, маппинг ошибок в HTTP-статусы, и smoke-тесты критичных хендлеров (health, register, login, создание комментария).

Намеренно не покрыто: `internal/repository` и `pkg/database` требуют реальной PostgreSQL — integration-тесты в рамках диплома не реализованы, SQL-запросы проверены вручную (`docker-compose up` + `curl`); `cmd/api` — точка входа, тоже проверена вручную. Остальные хендлеры (`PostHandler` целиком, часть `CommentHandler`) не покрыты — это дало бы прирост процента без новой информации о качестве кода, раз бизнес-логика под ними уже покрыта на уровне сервисов.

## Docker

### docker-compose.yml

Поднимает два сервиса:

- `db` (`blog_postgres_db`) — PostgreSQL 15 (`postgres:15-alpine`), данные — в volume `postgres_data`
- `app` (`blog_api`) — собранный Go-бинарник, стартует только после `service_healthy` у `db`

```bash
docker-compose up --build     # собрать и запустить все
docker-compose up -d db       # только БД, для локальной разработки
docker-compose logs -f app    # логи приложения
docker-compose logs -f db     # логи БД в реальном времени
docker-compose exec db psql -U postgres -d blog_db   # подключиться к БД
docker-compose down           # остановить
docker-compose down -v        # остановить и удалить данные БД
```

### Dockerfile

Многоступенчатая сборка: стадия `builder` (`golang:1.26-alpine`) компилирует бинарник; стадия `runtime` (`alpine:latest`) содержит только этот бинарник и папку `migrations` — без тулчейна, исходников и `go.mod`. Итоговый образ (`advanced-blog-management-system-app:latest` в выводе `docker images`) — около 11 МБ содержимого за счет этого.

## Известные ограничения

Проект соответствует требованиям дипломного задания; перечисленные ниже пункты — осознанные ограничения, а не недоделки:

- **Integration-тесты для repository не реализованы.** Требуют реальной PostgreSQL и оправданы для production-системы, но избыточны в рамках дипломного проекта
- **JWT нельзя отозвать до истечения TTL** (24 часа). Для production понадобится refresh-механизм или blacklist отозванных токенов
- **Нет rate-limiting на `/api/login` и `/api/register`** — оставляет пространство для brute-force атак на пароли
- **CORS настроен на `*`.** Для production следует сузить `Access-Control-Allow-Origin` до конкретных origin'ов
- **Отложенный лог хранится в одном файле.** Для нескольких инстансов приложения понадобится централизованное хранилище (например, stdout + внешний сборщик логов)

## Частые проблемы

**Порт 8080 уже занят.**
```bash
lsof -i :8080
docker-compose down
```

**БД не подключается.**
```bash
docker-compose ps
docker-compose logs db
```

**Тесты падают с ошибкой версии Go.**
```bash
go clean -cache
go clean -testcache
```

## Возможные доработки

- Refresh-токены для продления сессии без повторного логина
- Rate-limiting на чувствительные эндпоинты (login, register)
- Метрики Prometheus + Grafana для наблюдаемости
- Soft delete для постов и комментариев вместо физического удаления
- Полнотекстовый поиск по постам через PostgreSQL `tsvector`
- Integration-тесты репозиториев через `testcontainers-go`
- CI-пайплайн (GitHub Actions) с прогоном тестов, `go vet` и сборкой Docker-образа
