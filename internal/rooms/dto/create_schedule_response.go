package dto

// CreateScheduleResponse — ответ после создания расписания.
// @Description Ответ создания расписания
// @Name CreateScheduleResponse
type CreateScheduleResponse struct {
	Schedule *ScheduleJSON `json:"schedule"`
}
