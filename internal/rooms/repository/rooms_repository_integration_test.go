//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	bookingsmodel "github.com/internships-backend/test-backend-engory03/internal/bookings/model"
	"github.com/internships-backend/test-backend-engory03/internal/integrationtest"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/repository"
	usersmodel "github.com/internships-backend/test-backend-engory03/internal/users/model"
	"github.com/lib/pq"
)

func TestRoomRepository_Integration_CreateListExists(t *testing.T) {
	db := integrationtest.OpenPostgresAndMigrate(t)
	ctx := context.Background()
	rr := repository.NewRoomRepository(db)

	room := &model.Room{ID: uuid.New(), Name: "Integration Room"}
	if err := rr.Create(ctx, room); err != nil {
		t.Fatal(err)
	}
	ok, err := rr.ExistsByID(ctx, room.ID)
	if err != nil || !ok {
		t.Fatalf("ExistsByID: %v %v", ok, err)
	}
	list, err := rr.ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, r := range list {
		if r.ID == room.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("room not in ListAll")
	}
}

func TestScheduleRepository_Integration_CreateIfRoomExists(t *testing.T) {
	db := integrationtest.OpenPostgresAndMigrate(t)
	ctx := context.Background()
	roomRepo := repository.NewRoomRepository(db)
	schedRepo := repository.NewScheduleRepository(db)

	room := &model.Room{ID: uuid.New(), Name: "Sched Room"}
	if err := roomRepo.Create(ctx, room); err != nil {
		t.Fatal(err)
	}

	sched := &model.Schedule{
		ID:         uuid.New(),
		RoomID:     room.ID,
		DaysOfWeek: pq.Int64Array{1, 2},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}
	if err := schedRepo.CreateIfRoomExists(ctx, sched); err != nil {
		t.Fatal(err)
	}
	got, err := schedRepo.GetByRoomID(ctx, room.ID)
	if err != nil || got == nil || got.ID != sched.ID {
		t.Fatalf("GetByRoomID: %+v %v", got, err)
	}

	// duplicate room schedule → unique violation path
	sched2 := &model.Schedule{
		ID:         uuid.New(),
		RoomID:     room.ID,
		DaysOfWeek: pq.Int64Array{3},
		StartTime:  "11:00",
		EndTime:    "12:00",
	}
	err = schedRepo.CreateIfRoomExists(ctx, sched2)
	if err == nil {
		t.Fatal("expected duplicate schedule error")
	}
	if !errors.Is(err, repository.ErrScheduleDuplicateForRoom) {
		t.Fatalf("expected schedule duplicate error, got %v", err)
	}
}

func TestSlotRepository_Integration_BatchAndBookedLookup(t *testing.T) {
	db := integrationtest.OpenPostgresAndMigrate(t)
	ctx := context.Background()
	roomRepo := repository.NewRoomRepository(db)
	slotRepo := repository.NewSlotRepository(db)

	room := &model.Room{ID: uuid.New(), Name: "Slot Room"}
	if err := roomRepo.Create(ctx, room); err != nil {
		t.Fatal(err)
	}
	start := time.Now().UTC().Truncate(time.Minute).Add(48 * time.Hour)
	slots := []model.Slot{
		{ID: uuid.New(), RoomID: room.ID, StartAt: start, EndAt: start.Add(30 * time.Minute)},
		{ID: uuid.New(), RoomID: room.ID, StartAt: start.Add(30 * time.Minute), EndAt: start.Add(time.Hour)},
	}
	if err := slotRepo.UpsertGeneratedSlotsBatch(ctx, slots); err != nil {
		t.Fatal(err)
	}

	ids := []uuid.UUID{slots[0].ID, slots[1].ID}
	booked, err := slotRepo.ListActiveBookedSlotIDs(ctx, ids)
	if err != nil {
		t.Fatal(err)
	}
	if len(booked) != 0 {
		t.Fatalf("expected no bookings, got %v", booked)
	}

	uid := uuid.New()
	user := &usersmodel.User{ID: uid, Email: "slotbook@x", PasswordHash: "p", Role: "user"}
	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatal(err)
	}
	b := &bookingsmodel.Booking{
		ID:     uuid.New(),
		SlotID: slots[0].ID,
		UserID: uid,
		Status: "active",
	}
	if err := db.WithContext(ctx).Create(b).Error; err != nil {
		t.Fatal(err)
	}
	booked, err = slotRepo.ListActiveBookedSlotIDs(ctx, ids)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := booked[slots[0].ID]; !ok {
		t.Fatal("expected first slot booked")
	}
	if _, ok := booked[slots[1].ID]; ok {
		t.Fatal("second slot should not be booked")
	}
}
