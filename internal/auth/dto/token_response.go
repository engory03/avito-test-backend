package dto

// TokenResponse — ответ с одним JWT-токеном.
// @Description Ответ с JWT-токеном
// @Name TokenResponse
type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
