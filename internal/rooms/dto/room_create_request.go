package dto

// CreateRoomRequest — данные для создания переговорной.
// @Description Данные для создания переговорной
// @Name CreateRoomRequest
type CreateRoomRequest struct {
	Name        string  `json:"name" binding:"required" example:"Большая переговорная"`
	Description *string `json:"description" example:"Просторная переговорная с проектором и доской"`
	Capacity    *int    `json:"capacity" example:"20"`
}
