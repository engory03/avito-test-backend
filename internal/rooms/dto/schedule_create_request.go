package dto

import "github.com/google/uuid"

type ScheduleCreateRequest struct {
	RoomID     uuid.UUID `json:"roomId" binding:"required"`
	DaysOfWeek []int     `json:"daysOfWeek" binding:"required"`
	StartTime  string    `json:"startTime" binding:"required"`
	EndTime    string    `json:"endTime" binding:"required"`
}
