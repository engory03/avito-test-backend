package service

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
)

func TestValidateSchedulePayload_EmptyDays(t *testing.T) {
	err := validateSchedulePayload(&dto.ScheduleCreateRequest{
		RoomID: uuid.New(), DaysOfWeek: nil, StartTime: "09:00", EndTime: "10:00",
	})
	if err == nil || !strings.Contains(err.Error(), "daysOfWeek") {
		t.Fatalf("got %v", err)
	}
}

func TestValidateSchedulePayload_DayOutOfRange(t *testing.T) {
	err := validateSchedulePayload(&dto.ScheduleCreateRequest{
		RoomID: uuid.New(), DaysOfWeek: []int{8}, StartTime: "09:00", EndTime: "10:00",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateSchedulePayload_EndBeforeStart(t *testing.T) {
	err := validateSchedulePayload(&dto.ScheduleCreateRequest{
		RoomID: uuid.New(), DaysOfWeek: []int{1}, StartTime: "12:00", EndTime: "09:00",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateSchedulePayload_NotMultipleOf30(t *testing.T) {
	err := validateSchedulePayload(&dto.ScheduleCreateRequest{
		RoomID: uuid.New(), DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "09:45",
	})
	if err == nil || !strings.Contains(err.Error(), "30") {
		t.Fatalf("got %v", err)
	}
}

func TestValidateSchedulePayload_NormalizesTimes(t *testing.T) {
	req := &dto.ScheduleCreateRequest{
		RoomID: uuid.New(), DaysOfWeek: []int{1}, StartTime: "9:05", EndTime: "10:05",
	}
	if err := validateSchedulePayload(req); err != nil {
		t.Fatal(err)
	}
	if req.StartTime != "09:05" || req.EndTime != "10:05" {
		t.Fatalf("got %s %s", req.StartTime, req.EndTime)
	}
}

func TestParseHHMMToMinutes_FormatHHMM(t *testing.T) {
	m, err := parseHHMMToMinutes("14:30")
	if err != nil || m != 14*60+30 {
		t.Fatalf("%d %v", m, err)
	}
	if formatHHMM(m) != "14:30" {
		t.Fatal(formatHHMM(m))
	}
}
