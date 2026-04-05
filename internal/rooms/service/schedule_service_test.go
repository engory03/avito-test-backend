package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
)

type fakeScheduleRoomRepo struct {
	exists bool
	err    error
}

func (f *fakeScheduleRoomRepo) Create(context.Context, *model.Room) error { return nil }
func (f *fakeScheduleRoomRepo) ListAll(context.Context) ([]*model.Room, error) {
	return nil, nil
}
func (f *fakeScheduleRoomRepo) ExistsByID(context.Context, uuid.UUID) (bool, error) {
	return f.exists, f.err
}

type fakeScheduleRepo struct {
	exists bool
	err    error
	create func(context.Context, *model.Schedule) error
}

func (f *fakeScheduleRepo) Create(ctx context.Context, s *model.Schedule) error {
	if f.create != nil {
		return f.create(ctx, s)
	}
	return nil
}
func (f *fakeScheduleRepo) ExistsByRoomID(context.Context, uuid.UUID) (bool, error) {
	return f.exists, f.err
}
func (f *fakeScheduleRepo) GetByRoomID(context.Context, uuid.UUID) (*model.Schedule, error) {
	return nil, nil
}

func TestCreateSchedule_RoomIDMismatch(t *testing.T) {
	rid := uuid.New()
	other := uuid.New()
	svc := NewScheduleService(&fakeScheduleRoomRepo{}, &fakeScheduleRepo{})
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
	svc := NewScheduleService(&fakeScheduleRoomRepo{exists: false}, &fakeScheduleRepo{})
	_, err := svc.CreateSchedule(context.Background(), rid, dto.ScheduleCreateRequest{
		RoomID: rid, DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "10:00",
	})
	if !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestCreateSchedule_AlreadyExists(t *testing.T) {
	rid := uuid.New()
	svc := NewScheduleService(&fakeScheduleRoomRepo{exists: true}, &fakeScheduleRepo{exists: true})
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
	svc := NewScheduleService(&fakeScheduleRoomRepo{exists: true}, &fakeScheduleRepo{
		exists: false,
		create: func(_ context.Context, s *model.Schedule) error {
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
