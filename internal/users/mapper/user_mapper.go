package mapper

import (
	"github.com/internships-backend/test-backend-engory03/internal/users/dto"
	"github.com/internships-backend/test-backend-engory03/internal/users/model"
)

func NewUserResponse(u *model.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
