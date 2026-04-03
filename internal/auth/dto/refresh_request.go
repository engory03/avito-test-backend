package dto

// RefreshRequest — запрос обновления пары JWT-токенов.
// @Description Запрос обновления пары JWT-токенов
// @Name RefreshRequest
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required" example:"dGhpc2lzYXJlZnJlc2h0b2tlbg==..."`
}
