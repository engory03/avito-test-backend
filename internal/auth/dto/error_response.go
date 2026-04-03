package dto

// ErrorResponse — структура ответа с ошибкой аутентификации.
// @Description Ответ с описанием ошибки
// @Name ErrorResponse
type ErrorResponse struct {
	Message string `json:"message" example:"Неверные учётные данные"`
	Code    int    `json:"code" example:"401"`
}
