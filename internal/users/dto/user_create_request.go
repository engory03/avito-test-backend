package dto

// CreateUserRequest — создание пользователя администратором.
// @Description Запрос на создание пользователя
// @Name CreateUserRequest
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@mail.ru"`
	Password string `json:"password" binding:"required,min=6" example:"qwe123"`
	Role     string `json:"role" binding:"required,oneof=admin user" example:"admin"`
}
