package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
)

type ScheduleValidationError struct {
	Message string
}

func (e *ScheduleValidationError) Error() string {
	return e.Message
}

var scheduleTimePattern = regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)

func validateSchedulePayload(req *dto.ScheduleCreateRequest) error {
	if len(req.DaysOfWeek) == 0 {
		return &ScheduleValidationError{Message: "daysOfWeek must not be empty"}
	}
	for _, d := range req.DaysOfWeek {
		if d < 1 || d > 7 {
			return &ScheduleValidationError{Message: "each daysOfWeek value must be between 1 and 7"}
		}
	}
	if !scheduleTimePattern.MatchString(req.StartTime) {
		return &ScheduleValidationError{Message: "startTime must match HH:MM (00:00–23:59)"}
	}
	if !scheduleTimePattern.MatchString(req.EndTime) {
		return &ScheduleValidationError{Message: "endTime must match HH:MM (00:00–23:59)"}
	}
	startMin, err := parseHHMMToMinutes(req.StartTime)
	if err != nil {
		return &ScheduleValidationError{Message: "invalid startTime"}
	}
	endMin, err := parseHHMMToMinutes(req.EndTime)
	if err != nil {
		return &ScheduleValidationError{Message: "invalid endTime"}
	}
	if endMin <= startMin {
		return &ScheduleValidationError{Message: "endTime must be after startTime"}
	}
	if (endMin-startMin)%30 != 0 {
		return &ScheduleValidationError{Message: "working window duration must be a multiple of 30 minutes for 30-minute slots"}
	}
	req.StartTime = formatHHMM(startMin)
	req.EndTime = formatHHMM(endMin)
	return nil
}

func parseHHMMToMinutes(s string) (int, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time")
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, fmt.Errorf("out of range")
	}
	return h*60 + m, nil
}

func formatHHMM(minutes int) string {
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}
