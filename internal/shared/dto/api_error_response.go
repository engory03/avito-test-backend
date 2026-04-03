package shareddto

// APIErrorDetail Детали ошибки: код и сообщение
// @Description Тело ошибки (.error)
// @Name APIErrorDetail
type APIErrorDetail struct {
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"invalid request"`
}

// APIErrorResponse Стандартный контейнер ошибки из api.yaml (ErrorResponse / InternalErrorResponse).
// @Description Стандартный контейнер ошибки из api.yaml (ErrorResponse / InternalErrorResponse).
// @Name APIErrorResponse
type APIErrorResponse struct {
	Error APIErrorDetail `json:"error"`
}
