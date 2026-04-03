package dto

import (
	"time"

	"github.com/google/uuid"
)

// UserResponse — данные пользователя в ответе API.
// @Description Ответ с данными пользователя
// @Name UserResponse
type UserResponse struct {
	ID        uuid.UUID `json:"id" example:"a8098c1a-f86e-11da-bd1a-00112444be1e"`
	Email     string    `json:"email" example:"ivanov@example.com"`
	Role      string    `json:"role" example:"admin"`
	CreatedAt time.Time `json:"createdAt" example:"2025-11-12T19:45:00Z"`
	UpdatedAt time.Time `json:"updatedAt" example:"2025-11-12T19:45:00Z"`
}
