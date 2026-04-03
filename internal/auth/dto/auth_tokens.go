package dto

// AuthTokens представляет JWT-токены.
// @Description Ответ с JWT-токенами
// @Name AuthTokens
type AuthTokens struct {
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refreshToken" example:"dGhpc2lzYXJlZnJlc2h0b2tlbg==..."`
}
