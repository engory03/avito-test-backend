package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/repository"
)

type fakeScheduleRepo struct {
	createIfRoomExists func(context.Context, *model.Schedule) error
}

func (f *fakeScheduleRepo) CreateIfRoomExists(ctx context.Context, s *model.Schedule) error {
	if f.createIfRoomExists != nil {
		return f.createIfRoomExists(ctx, s)
	}
	return nil
}
func (f *fakeScheduleRepo) ExistsByRoomID(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (f *fakeScheduleRepo) GetByRoomID(context.Context, uuid.UUID) (*model.Schedule, error) {
	return nil, nil
}

func TestCreateSchedule_RoomIDMismatch(t *testing.T) {
	rid := uuid.New()
	other := uuid.New()
	svc := NewScheduleService(&fakeScheduleRepo{})
	_, err := svc.CreateSchedule(context.Background(), rid, dto.ScheduleCreateRequest{
		RoomID: other, DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "10:00",
	})
	var ve *ScheduleValidationError
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.As(err, &ve) {
		t.Fatalf("got %T %v", err, err)
	}
}

func TestCreateSchedule_RoomNotFound(t *testing.T) {
	rid := uuid.New()
	svc := NewScheduleService(&fakeScheduleRepo{
		createIfRoomExists: func(context.Context, *model.Schedule) error {
			return repository.ErrScheduleRoomNotFound
		},
	})
	_, err := svc.CreateSchedule(context.Background(), rid, dto.ScheduleCreateRequest{
		RoomID: rid, DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "10:00",
	})
	if !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestCreateSchedule_AlreadyExists(t *testing.T) {
	rid := uuid.New()
	svc := NewScheduleService(&fakeScheduleRepo{
		createIfRoomExists: func(context.Context, *model.Schedule) error {
			return repository.ErrScheduleDuplicateForRoom
		},
	})
	_, err := svc.CreateSchedule(context.Background(), rid, dto.ScheduleCreateRequest{
		RoomID: rid, DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "10:00",
	})
	if !errors.Is(err, ErrScheduleExists) {
		t.Fatalf("got %v", err)
	}
}

func TestCreateSchedule_Success(t *testing.T) {
	rid := uuid.New()
	var saved *model.Schedule
	svc := NewScheduleService(&fakeScheduleRepo{
		createIfRoomExists: func(_ context.Context, s *model.Schedule) error {
			saved = s
			return nil
		},
	})
	got, err := svc.CreateSchedule(context.Background(), rid, dto.ScheduleCreateRequest{
		RoomID: rid, DaysOfWeek: []int{1, 3}, StartTime: "9:00", EndTime: "10:00",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.RoomID != rid || len(got.DaysOfWeek) != 2 {
		t.Fatalf("got %+v", got)
	}
	if saved == nil || saved.StartTime != "09:00" || saved.EndTime != "10:00" {
		t.Fatalf("saved %+v", saved)
	}
	if len(got.DaysOfWeek) < 1 || got.DaysOfWeek[0] != 1 {
		t.Fatalf("days %+v", got.DaysOfWeek)
	}
}
