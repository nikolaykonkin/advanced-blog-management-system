package service

import (
	"context"
	"errors"
	"testing"

	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"
	"advanced-blog-management-system/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserService_Register_Success(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	user, err := svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "user1",
		Email:    "user1@test.com",
		Password: "password123",
	})

	require.NoError(t, err)
	assert.Equal(t, "user1", user.Username)
	assert.Equal(t, "user1@test.com", user.Email)
	assert.NotEqual(t, "password123", user.Password)
	assert.True(t, auth.VerifyPassword(user.Password, "password123"))
}

func TestUserService_Register_DuplicateEmail_ReturnsErrUserAlreadyExists(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	_, err := svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "user1", Email: "taken@test.com", Password: "password123",
	})
	require.NoError(t, err)

	_, err = svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "user2", Email: "taken@test.com", Password: "password123",
	})

	assert.ErrorIs(t, err, apperrors.ErrUserAlreadyExists)
}

func TestUserService_Register_DuplicateUsername_ReturnsErrUserAlreadyExists(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	_, err := svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "taken", Email: "user1@test.com", Password: "password123",
	})
	require.NoError(t, err)

	_, err = svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "taken", Email: "user2@test.com", Password: "password123",
	})

	assert.ErrorIs(t, err, apperrors.ErrUserAlreadyExists)
}

func TestUserService_Register_InvalidRequest_ReturnsValidationError(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	_, err := svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "user1", Email: "not-an-email", Password: "password123",
	})

	assert.Error(t, err)
}

func TestUserService_Register_RepoCreateDuplicateError_ReturnsErrUserAlreadyExists(t *testing.T) {
	repo := newFakeUserRepository()
	repo.createErr = repository.ErrDuplicateUser
	svc := NewUserService(repo)

	_, err := svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "user1", Email: "user1@test.com", Password: "password123",
	})

	assert.ErrorIs(t, err, apperrors.ErrUserAlreadyExists)
}

func TestUserService_Login_Success(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	_, err := svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "user1", Email: "user1@test.com", Password: "password123",
	})
	require.NoError(t, err)

	user, err := svc.Login(context.Background(), &model.UserLoginRequest{
		Email: "user1@test.com", Password: "password123",
	})

	require.NoError(t, err)
	assert.Equal(t, "user1@test.com", user.Email)
}

func TestUserService_Login_UserNotFound_ReturnsErrInvalidCredentials(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	_, err := svc.Login(context.Background(), &model.UserLoginRequest{
		Email: "ghost@test.com", Password: "password123",
	})

	assert.ErrorIs(t, err, apperrors.ErrInvalidCredentials)
}

func TestUserService_Login_WrongPassword_ReturnsErrInvalidCredentials(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	_, err := svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "user1", Email: "user1@test.com", Password: "password123",
	})
	require.NoError(t, err)

	_, err = svc.Login(context.Background(), &model.UserLoginRequest{
		Email: "user1@test.com", Password: "wrongpassword",
	})

	// "Пользователь не найден" и "неверный пароль" должны давать одну и
	// ту же ошибку - защита от user enumeration
	assert.ErrorIs(t, err, apperrors.ErrInvalidCredentials)
}

func TestUserService_GetByID_NotFound_ReturnsErrUserNotFound(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	_, err := svc.GetByID(context.Background(), 999)

	assert.ErrorIs(t, err, apperrors.ErrUserNotFound)
}

func TestUserService_GetByEmail_NotFound_ReturnsErrUserNotFound(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	_, err := svc.GetByEmail(context.Background(), "ghost@test.com")

	assert.ErrorIs(t, err, apperrors.ErrUserNotFound)
}

func TestUserService_Register_ExistsByEmailError_ReturnsError(t *testing.T) {
	repo := newFakeUserRepository()
	repo.existsByEmailErr = errors.New("db unavailable")
	svc := NewUserService(repo)

	_, err := svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "user1", Email: "user1@test.com", Password: "password123",
	})

	assert.Error(t, err)
}

func TestUserService_Register_ExistsByUsernameError_ReturnsError(t *testing.T) {
	repo := newFakeUserRepository()
	repo.existsByUsernameErr = errors.New("db unavailable")
	svc := NewUserService(repo)

	_, err := svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "user1", Email: "user1@test.com", Password: "password123",
	})

	assert.Error(t, err)
}

// в отличие от TestUserService_Register_RepoCreateDuplicateError_ReturnsErrUserAlreadyExists,
// здесь ошибка Create - НЕ repository.ErrDuplicateUser, и не должна маппиться в ErrUserAlreadyExists
func TestUserService_Register_CreateGenericError_ReturnsError(t *testing.T) {
	repo := newFakeUserRepository()
	repo.createErr = errors.New("disk full")
	svc := NewUserService(repo)

	_, err := svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "user1", Email: "user1@test.com", Password: "password123",
	})

	assert.Error(t, err)
	assert.NotErrorIs(t, err, apperrors.ErrUserAlreadyExists)
}

func TestUserService_Login_GetByEmailError_ReturnsError(t *testing.T) {
	repo := newFakeUserRepository()
	repo.getByEmailErr = errors.New("db unavailable")
	svc := NewUserService(repo)

	_, err := svc.Login(context.Background(), &model.UserLoginRequest{
		Email: "user1@test.com", Password: "password123",
	})

	assert.Error(t, err)
}

func TestUserService_GetByID_Success(t *testing.T) {
	repo := newFakeUserRepository()
	svc := NewUserService(repo)
	created, err := svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "user1", Email: "user1@test.com", Password: "password123",
	})
	require.NoError(t, err)

	user, err := svc.GetByID(context.Background(), created.ID)

	require.NoError(t, err)
	assert.Equal(t, "user1@test.com", user.Email)
}

func TestUserService_GetByID_RepoError_ReturnsError(t *testing.T) {
	repo := newFakeUserRepository()
	repo.getByIDErr = errors.New("db unavailable")
	svc := NewUserService(repo)

	_, err := svc.GetByID(context.Background(), 1)

	assert.Error(t, err)
}

func TestUserService_GetByEmail_Success(t *testing.T) {
	repo := newFakeUserRepository()
	svc := NewUserService(repo)
	_, err := svc.Register(context.Background(), &model.UserCreateRequest{
		Username: "user1", Email: "user1@test.com", Password: "password123",
	})
	require.NoError(t, err)

	user, err := svc.GetByEmail(context.Background(), "user1@test.com")

	require.NoError(t, err)
	assert.Equal(t, "user1", user.Username)
}

func TestUserService_GetByEmail_RepoError_ReturnsError(t *testing.T) {
	repo := newFakeUserRepository()
	repo.getByEmailErr = errors.New("db unavailable")
	svc := NewUserService(repo)

	_, err := svc.GetByEmail(context.Background(), "user1@test.com")

	assert.Error(t, err)
}
