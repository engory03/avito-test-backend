package dto

// LoginRequest represents user login data.
// @Description Login request payload
// @Name LoginRequest
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@mail.ru"`
	Password string `json:"password" binding:"required" example:"qwe123"`
}
