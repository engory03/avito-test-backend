package dto

import "github.com/google/uuid"

// ScheduleJSON — расписание переговорной в JSON.
// @Description Данные расписания
// @Name ScheduleJSON
type ScheduleJSON struct {
	ID         uuid.UUID `json:"id" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	RoomID     uuid.UUID `json:"roomId" example:"a8098c1a-f86e-11da-bd1a-00112444be1e"`
	DaysOfWeek []int     `json:"daysOfWeek" example:"1,3,5"`
	StartTime  string    `json:"startTime" example:"09:00"`
	EndTime    string    `json:"endTime" example:"18:00"`
}
