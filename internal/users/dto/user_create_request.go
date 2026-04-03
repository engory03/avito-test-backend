package dto

// CreateUserRequest represents a user registration request.
// @Description User registration request payload
// @Name CreateUserRequest
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email" example:"ivanov@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"qwe123"`
	Role     string `json:"role" binding:"required,oneof=admin user" example:"admin"`
}
