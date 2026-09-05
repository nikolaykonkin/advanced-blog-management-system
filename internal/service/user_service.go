package service

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"
	"advanced-blog-management-system/pkg/auth"
	"context"
	"errors"
	"fmt"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, req *model.UserCreateRequest) (*model.User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	emailExists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if emailExists {
		return nil, apperrors.ErrUserAlreadyExists
	}

	usernameExists, err := s.repo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if usernameExists {
		return nil, apperrors.ErrUserAlreadyExists
	}

	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateUser) {
			return nil, apperrors.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, req *model.UserLoginRequest) (*model.User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	if user == nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	if !auth.VerifyPassword(user.Password, req.Password) {
		return nil, apperrors.ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id int) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}
	return user, nil
}
