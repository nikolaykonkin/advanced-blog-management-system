package service

import (
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"
	"context"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, req *model.UserCreateRequest) (*model.User, error) {
	// TODO: Реализовать регистрацию пользователя
	// 1. Валидировать входные данные (req.Validate())
	// 2. Проверить что пользователь с таким email не существует
	// 3. Проверить что пользователь с таким username не существует
	// 4. Захешировать пароль (использовать golang.org/x/crypto/bcrypt)
	// 5. Создать пользователя через репозиторий
	// 6. Вернуть созданного пользователя
	return nil, nil
}

func (s *UserService) Login(ctx context.Context, req *model.UserLoginRequest) (*model.User, error) {
	// TODO: Реализовать вход пользователя
	// 1. Валидировать входные данные
	// 2. Получить пользователя по email
	// 3. Если не найден - вернуть ошибку
	// 4. Сравнить пароль с хешем (bcrypt.CompareHashAndPassword)
	// 5. Вернуть пользователя
	return nil, nil
}

func (s *UserService) GetByID(ctx context.Context, id int) (*model.User, error) {
	// TODO: Реализовать получение пользователя по ID
	return nil, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	// TODO: Реализовать получение пользователя по email
	return nil, nil
}
