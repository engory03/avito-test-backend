package mapper

import (
	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
)

func ScheduleToJSON(s *model.Schedule) *dto.ScheduleJSON {
	if s == nil {
		return nil
	}
	return &dto.ScheduleJSON{
		ID:         s.ID,
		RoomID:     s.RoomID,
		DaysOfWeek: s.DaysOfWeek,
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
	}
}
