package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
	"github.com/internships-backend/test-backend-engory03/internal/users/dto"
	"github.com/internships-backend/test-backend-engory03/internal/users/model"
	"github.com/internships-backend/test-backend-engory03/internal/users/repository"
)

type UserService interface {
	CreateUser(req dto.CreateUserRequest) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetUserById(id uuid.UUID) (*model.User, error)
	GetAllUsers() ([]*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userService{repo: r}
}

func (s *userService) CreateUser(req dto.CreateUserRequest) (*model.User, error) {
	// 1. Check if email already exists
	existingUser, err := s.repo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// 2. Hash password before storing
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

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetByEmail(email string) (*model.User, error) {
	return s.repo.GetByEmail(email)
}

func (s *userService) GetUserById(id uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) GetAllUsers() ([]*model.User, error) {
	users, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return users, nil
}
