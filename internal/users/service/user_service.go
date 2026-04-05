package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
	"github.com/internships-backend/test-backend-engory03/internal/users/dto"
	"github.com/internships-backend/test-backend-engory03/internal/users/model"
	"github.com/internships-backend/test-backend-engory03/internal/users/repository"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserById(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetAllUsers(ctx context.Context) ([]*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userService{repo: r}
}

func (s *userService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*model.User, error) {
	existingUser, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &model.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *userService) GetUserById(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	users, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
