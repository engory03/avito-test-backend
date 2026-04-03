package shareddto

// ErrorDefaultResponse represents the structure of an error response.
// @Description Error response payload
// @Name ErrorDefaultResponse
type ErrorDefaultResponse struct {
	Error string `json:"error" example:"error message"`
}
