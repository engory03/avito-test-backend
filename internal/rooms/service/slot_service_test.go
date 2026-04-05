package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"github.com/lib/pq"
)

type fakeSlotRoomRepo struct {
	exists bool
	err    error
}

func (f *fakeSlotRoomRepo) Create(context.Context, *model.Room) error { return nil }
func (f *fakeSlotRoomRepo) ListAll(context.Context) ([]*model.Room, error) {
	return nil, nil
}
func (f *fakeSlotRoomRepo) ExistsByID(context.Context, uuid.UUID) (bool, error) {
	return f.exists, f.err
}

type fakeSlotScheduleRepo struct {
	sched *model.Schedule
	err   error
}

func (f *fakeSlotScheduleRepo) CreateIfRoomExists(context.Context, *model.Schedule) error { return nil }
func (f *fakeSlotScheduleRepo) ExistsByRoomID(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (f *fakeSlotScheduleRepo) GetByRoomID(context.Context, uuid.UUID) (*model.Schedule, error) {
	return f.sched, f.err
}

type fakeSlotRepo struct {
	upsertErr error
	booked    map[uuid.UUID]struct{}
}

func (f *fakeSlotRepo) UpsertGeneratedSlotsBatch(context.Context, []model.Slot) error {
	return f.upsertErr
}

func (f *fakeSlotRepo) ListActiveBookedSlotIDs(_ context.Context, slotIDs []uuid.UUID) (map[uuid.UUID]struct{}, error) {
	if f.booked != nil {
		return f.booked, nil
	}
	return map[uuid.UUID]struct{}{}, nil
}

func TestListAvailableSlots_InvalidDate(t *testing.T) {
	svc := NewSlotService(&fakeSlotRoomRepo{}, &fakeSlotScheduleRepo{}, &fakeSlotRepo{})
	_, err := svc.ListAvailableSlots(context.Background(), uuid.New(), "")
	if !errors.Is(err, ErrInvalidSlotListDate) {
		t.Fatalf("got %v", err)
	}
	_, err = svc.ListAvailableSlots(context.Background(), uuid.New(), "not-a-date")
	if err == nil || !errors.Is(err, ErrInvalidSlotListDate) {
		t.Fatalf("got %v", err)
	}
}

func TestListAvailableSlots_RoomNotFound(t *testing.T) {
	svc := NewSlotService(&fakeSlotRoomRepo{exists: false}, &fakeSlotScheduleRepo{}, &fakeSlotRepo{})
	_, err := svc.ListAvailableSlots(context.Background(), uuid.New(), "2026-04-06")
	if !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestListAvailableSlots_NoSchedule(t *testing.T) {
	rid := uuid.New()
	svc := NewSlotService(&fakeSlotRoomRepo{exists: true}, &fakeSlotScheduleRepo{sched: nil}, &fakeSlotRepo{})
	got, err := svc.ListAvailableSlots(context.Background(), rid, "2026-04-06")
	if err != nil || got != nil {
		t.Fatalf("got %v err %v", got, err)
	}
}

func TestListAvailableSlots_WeekdayNotInSchedule(t *testing.T) {
	rid := uuid.New()
	sched := &model.Schedule{
		RoomID: rid, DaysOfWeek: pq.Int64Array{2},
		StartTime: "09:00", EndTime: "18:00",
	}
	svc := NewSlotService(&fakeSlotRoomRepo{exists: true}, &fakeSlotScheduleRepo{sched: sched}, &fakeSlotRepo{})
	got, err := svc.ListAvailableSlots(context.Background(), rid, "2026-04-06")
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v err %v", got, err)
	}
}

func TestListAvailableSlots_GeneratesFreeSlots(t *testing.T) {
	rid := uuid.New()
	sched := &model.Schedule{
		RoomID: rid, DaysOfWeek: pq.Int64Array{1},
		StartTime: "09:00", EndTime: "10:00",
	}
	svc := NewSlotService(&fakeSlotRoomRepo{exists: true}, &fakeSlotScheduleRepo{sched: sched}, &fakeSlotRepo{})
	got, err := svc.ListAvailableSlots(context.Background(), rid, "2026-04-06")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 slots got %d %+v", len(got), got)
	}
	if !got[0].StartAt.Before(got[0].EndAt) {
		t.Fatal("slot bounds")
	}
}

func TestListAvailableSlots_FiltersBooked(t *testing.T) {
	rid := uuid.New()
	sched := &model.Schedule{
		RoomID: rid, DaysOfWeek: pq.Int64Array{1},
		StartTime: "09:00", EndTime: "10:00",
	}
	svc := NewSlotService(&fakeSlotRoomRepo{exists: true}, &fakeSlotScheduleRepo{sched: sched}, &fakeSlotRepo{
		booked: nil,
	})
	all, err := svc.ListAvailableSlots(context.Background(), rid, "2026-04-06")
	if err != nil || len(all) != 2 {
		t.Fatalf("setup %+v", all)
	}
	firstID := all[0].ID
	svc2 := NewSlotService(&fakeSlotRoomRepo{exists: true}, &fakeSlotScheduleRepo{sched: sched}, &fakeSlotRepo{
		booked: map[uuid.UUID]struct{}{firstID: {}},
	})
	got, err := svc2.ListAvailableSlots(context.Background(), rid, "2026-04-06")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID == firstID {
		t.Fatalf("got %+v", got)
	}
}

func TestListAvailableSlots_InvalidScheduleTime(t *testing.T) {
	rid := uuid.New()
	sched := &model.Schedule{
		RoomID: rid, DaysOfWeek: pq.Int64Array{1},
		StartTime: "25:00", EndTime: "26:00",
	}
	svc := NewSlotService(&fakeSlotRoomRepo{exists: true}, &fakeSlotScheduleRepo{sched: sched}, &fakeSlotRepo{})
	_, err := svc.ListAvailableSlots(context.Background(), rid, "2026-04-06")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestListAvailableSlots_WindowNotPositive(t *testing.T) {
	rid := uuid.New()
	sched := &model.Schedule{
		RoomID: rid, DaysOfWeek: pq.Int64Array{1},
		StartTime: "12:00", EndTime: "12:00",
	}
	svc := NewSlotService(&fakeSlotRoomRepo{exists: true}, &fakeSlotScheduleRepo{sched: sched}, &fakeSlotRepo{})
	got, err := svc.ListAvailableSlots(context.Background(), rid, "2026-04-06")
	if err != nil || got != nil {
		t.Fatalf("got %v err %v", got, err)
	}
}

func TestApiWeekday_SundayIsSeven(t *testing.T) {
	sun := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	if apiWeekday(sun) != 7 {
		t.Fatalf("got %d", apiWeekday(sun))
	}
}
