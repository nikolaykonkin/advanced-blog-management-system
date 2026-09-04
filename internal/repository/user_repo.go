package repository

import (
	"advanced-blog-management-system/internal/model"
	"context"
	"database/sql"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	// TODO: Реализовать создание пользователя
	// Вставить пользователя в таблицу users
	// Вернуть ошибку, если email или username уже существуют
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	// TODO: Реализовать получение пользователя по ID
	// Выполнить SELECT запрос
	// Вернуть nil, nil если пользователь не найден
	return nil, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	// TODO: Реализовать получение пользователя по email
	return nil, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	// TODO: Реализовать получение пользователя по username
	return nil, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	// TODO: Реализовать проверку существования по email
	return false, nil
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	// TODO: Реализовать проверку существования по username
	return false, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	// TODO: Реализовать обновление пользователя
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	// TODO: Реализовать удаление пользователя
	return nil
}
