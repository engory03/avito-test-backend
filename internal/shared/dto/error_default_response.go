package shareddto

// ErrorDefaultResponse Структура ответа с ошибкой
// @Description Ответ с ошибкой
// @Name ErrorDefaultResponse
type ErrorDefaultResponse struct {
	Error string `json:"error" example:"текст ошибки"`
}
