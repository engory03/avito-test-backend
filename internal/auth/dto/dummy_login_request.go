package dto

// DummyLoginRequest — тестовый вход по роли (без пароля).
// @Description Запрос фиктивного входа по роли
// @Name DummyLoginRequest
type DummyLoginRequest struct {
	Role string `json:"role" binding:"required,oneof=admin user" example:"admin"`
}
