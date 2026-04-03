package dto

// LoginRequest — данные для входа в систему.
// @Description Запрос на вход
// @Name LoginRequest
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@mail.ru"`
	Password string `json:"password" binding:"required" example:"qwe123"`
}
