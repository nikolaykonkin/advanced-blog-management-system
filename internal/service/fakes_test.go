package service

import (
	"context"

	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"
)

// fakeUserRepository - минимальная in-memory реализация repository.UserRepository
// для юнит-тестов UserService, без обращения к реальной базе данных
type fakeUserRepository struct {
	byEmail    map[string]*model.User
	byUsername map[string]*model.User
	byID       map[int]*model.User
	nextID     int

	// createErr, если задан, возвращается из Create вместо реальной вставки -
	// используется для симуляции гонки на уровне БД (два параллельных запроса
	// проходят проверку ExistsByEmail/Username, но Create всё равно падает с ErrDuplicateUser)
	createErr error
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		byEmail:    make(map[string]*model.User),
		byUsername: make(map[string]*model.User),
		byID:       make(map[int]*model.User),
	}
}

func (r *fakeUserRepository) Create(ctx context.Context, user *model.User) error {
	if r.createErr != nil {
		return r.createErr
	}

	r.nextID++
	user.ID = r.nextID
	r.byEmail[user.Email] = user
	r.byUsername[user.Username] = user
	r.byID[user.ID] = user
	return nil
}

func (r *fakeUserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	return r.byID[id], nil
}

func (r *fakeUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.byEmail[email], nil
}

func (r *fakeUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return r.byUsername[username], nil
}

func (r *fakeUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, ok := r.byEmail[email]
	return ok, nil
}

func (r *fakeUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	_, ok := r.byUsername[username]
	return ok, nil
}

func (r *fakeUserRepository) Update(ctx context.Context, user *model.User) error {
	r.byID[user.ID] = user
	return nil
}

func (r *fakeUserRepository) Delete(ctx context.Context, id int) error {
	delete(r.byID, id)
	return nil
}

// Компиляционная проверка: fakeUserRepository должен реализовывать весь интерфейс
// repository.UserRepository, а не только методы, используемые в текущих тестах
var _ repository.UserRepository = (*fakeUserRepository)(nil)
