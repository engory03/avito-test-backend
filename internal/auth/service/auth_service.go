package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/auth/dto"
	"github.com/internships-backend/test-backend-engory03/internal/auth/dummy"
	"github.com/internships-backend/test-backend-engory03/internal/auth/model"
	"github.com/internships-backend/test-backend-engory03/internal/auth/repository"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
	usersModel "github.com/internships-backend/test-backend-engory03/internal/users/model"
	usersRepo "github.com/internships-backend/test-backend-engory03/internal/users/repository"
)

var ErrInvalidDummyRole = errors.New("invalid role for dummy login")

type AuthService interface {
	Register(req dto.RegisterRequest) (*dto.AuthTokens, error)
	Login(req dto.LoginRequest) (*dto.AuthTokens, error)
	Refresh(refreshToken string) (*dto.AuthTokens, error)
	Logout(refreshToken string) error
	DummyLogin(req dto.DummyLoginRequest) (string, error)
}

type authService struct {
	userRepo    usersRepo.UserRepository
	refreshRepo repository.RefreshTokenRepository
	jwt         *utils.JWTManager
}

func NewAuthService(u usersRepo.UserRepository, r repository.RefreshTokenRepository, j *utils.JWTManager) AuthService {
	return &authService{userRepo: u, refreshRepo: r, jwt: j}
}

func (s *authService) Register(req dto.RegisterRequest) (*dto.AuthTokens, error) {
	// 1. Check if user already exists
	existing, _ := s.userRepo.GetByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("user with this email already exists")
	}

	// 2. Hash password
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 3. Create user model
	user := &usersModel.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashed,
		Role:         req.Role,
	}

	// 4. Save to DB
	if createErr := s.userRepo.Create(user); createErr != nil {
		return nil, createErr
	}

	// 5. Generate JWTs
	access, err := s.jwt.GenerateAccessToken(user.ID, "user")
	if err != nil {
		return nil, err
	}

	refresh, err := s.jwt.GenerateRefreshToken(user.ID, "user")
	if err != nil {
		return nil, err
	}

	// 6. Store refresh token
	tokenModel := &model.RefreshToken{
		UserID:    user.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.refreshRepo.Save(tokenModel); err != nil {
		return nil, err
	}

	return &dto.AuthTokens{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.AuthTokens, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	access, err := s.jwt.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	refresh, err := s.jwt.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	tokenModel := &model.RefreshToken{
		UserID:    user.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.refreshRepo.Save(tokenModel); err != nil {
		return nil, err
	}

	return &dto.AuthTokens{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *authService) Refresh(refreshToken string) (*dto.AuthTokens, error) {
	claims, err := s.jwt.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	role := claims.Role

	access, err := s.jwt.GenerateAccessToken(userID, role)
	if err != nil {
		return nil, err
	}

	newRefresh, err := s.jwt.GenerateRefreshToken(userID, role)
	if err != nil {
		return nil, err
	}

	_ = s.refreshRepo.Delete(refreshToken)

	tokenModel := &model.RefreshToken{
		UserID:    userID,
		Token:     newRefresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.refreshRepo.Save(tokenModel); err != nil {
		return nil, err
	}

	return &dto.AuthTokens{
		AccessToken:  access,
		RefreshToken: newRefresh,
	}, nil
}

func (s *authService) Logout(refreshToken string) error {
	return s.refreshRepo.Delete(refreshToken)
}

func (s *authService) DummyLogin(req dto.DummyLoginRequest) (string, error) {
	var uid uuid.UUID
	switch req.Role {
	case "admin":
		uid = dummy.DummyAdminID
	case "user":
		uid = dummy.DummyUserID
	default:
		return "", ErrInvalidDummyRole
	}
	return s.jwt.GenerateAccessToken(uid, req.Role)
}
