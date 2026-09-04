package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgresDB создает новое подключение к PostgreSQL
func NewPostgresDB(cfg Config) (*sql.DB, error) {
	// TODO: Реализовать подключение к PostgreSQL
	// 1. Создать DSN (Data Source Name) строку в формате:
	//    host=<host> port=<port> user=<user> password=<password> dbname=<dbname> sslmode=<sslmode>
	// 2. Открыть подключение используя sql.Open("postgres", dsn)
	// 3. Проверить подключение используя db.Ping()
	// 4. Если ошибка - вернуть ошибку
	// 5. Вернуть подключение (*sql.DB)
	// 6. (Рекомендуется) Установить параметры пула: SetMaxOpenConns(25), SetMaxIdleConns(5)

	return nil, nil
}

// RunMigrations выполняет SQL миграции
func RunMigrations(db *sql.DB, migrations []string) error {
	// TODO: Реализовать выполнение миграций
	// 1. Для каждой миграции в списке:
	//    - Выполнить миграцию используя db.Exec()
	//    - Если ошибка - залогировать и вернуть ошибку
	// 2. Вернуть nil если все миграции прошли успешно

	return nil
}
