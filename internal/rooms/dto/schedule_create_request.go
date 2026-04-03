package dto

import "github.com/google/uuid"

// ScheduleCreateRequest — данные для создания расписания.
// @Description Данные для создания расписания
// @Name ScheduleCreateRequest
type ScheduleCreateRequest struct {
	RoomID     uuid.UUID `json:"roomId" binding:"required" example:"a8098c1a-f86e-11da-bd1a-00112444be1e"`
	DaysOfWeek []int     `json:"daysOfWeek" binding:"required" example:"1,3,5"`
	StartTime  string    `json:"startTime" binding:"required" example:"09:00"`
	EndTime    string    `json:"endTime" binding:"required" example:"18:00"`
}
