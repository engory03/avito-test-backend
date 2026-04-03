package dto

// ErrorResponse represents the structure of an error response.
// @Description Error response payload
// @Name ErrorResponse
type ErrorResponse struct {
	Message string `json:"message" example:"Invalid credentials"`
	Code    int    `json:"code" example:"401"`
}
