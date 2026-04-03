package shareddto

// APIErrorDetail matches components.schemas.ErrorResponse / InternalErrorResponse (.error).
// @Description Error payload (.error)
// @Name APIErrorDetail
type APIErrorDetail struct {
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"invalid request"`
}

// APIErrorResponse is the standard error envelope from api.yaml (ErrorResponse / InternalErrorResponse).
// @Description Standard API error response
// @Name APIErrorResponse
type APIErrorResponse struct {
	Error APIErrorDetail `json:"error"`
}
