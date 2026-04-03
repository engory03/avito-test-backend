package dto

// RegisterRequest — тело запроса регистрации пользователя.
// @Description Запрос на регистрацию
// @Name RegisterRequest
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@mail.ru"`
	Password string `json:"password" binding:"required,min=6" example:"qwe123"`
	Role     string `json:"role" binding:"required,oneof=admin user" example:"admin"`
}
