//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/model"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/repository"
	"github.com/internships-backend/test-backend-engory03/internal/integrationtest"
	roomsmodel "github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	roomsrepo "github.com/internships-backend/test-backend-engory03/internal/rooms/repository"
	usersmodel "github.com/internships-backend/test-backend-engory03/internal/users/model"
	usersrepo "github.com/internships-backend/test-backend-engory03/internal/users/repository"
)

func TestBookingRepository_Integration_CreateLockAndCancel(t *testing.T) {
	db, cleanup := integrationtest.OpenPostgresAndMigrate(t)
	t.Cleanup(cleanup)
	ctx := context.Background()

	ur := usersrepo.NewUserRepository(db)
	user := &usersmodel.User{ID: uuid.New(), Email: "book-int@x", PasswordHash: "h", Role: "user"}
	if err := ur.Create(ctx, user); err != nil {
		t.Fatal(err)
	}

	rr := roomsrepo.NewRoomRepository(db)
	room := &roomsmodel.Room{ID: uuid.New(), Name: "Book Room"}
	if err := rr.Create(ctx, room); err != nil {
		t.Fatal(err)
	}

	start := time.Now().UTC().Truncate(time.Minute).Add(72 * time.Hour)
	slot := &roomsmodel.Slot{
		ID:      uuid.New(),
		RoomID:  room.ID,
		StartAt: start,
		EndAt:   start.Add(30 * time.Minute),
	}
	if err := db.WithContext(ctx).Create(slot).Error; err != nil {
		t.Fatal(err)
	}

	br := repository.NewBookingRepository(db)
	now := time.Now().UTC()
	b1 := &model.Booking{
		ID:     uuid.New(),
		SlotID: slot.ID,
		UserID: user.ID,
		Status: "active",
	}
	if err := br.CreateActiveBookingWithSlotLock(ctx, b1, slot.ID, now); err != nil {
		t.Fatal(err)
	}

	b2 := &model.Booking{
		ID:     uuid.New(),
		SlotID: slot.ID,
		UserID: user.ID,
		Status: "active",
	}
	err := br.CreateActiveBookingWithSlotLock(ctx, b2, slot.ID, now)
	if !errors.Is(err, repository.ErrActiveBookingExists) {
		t.Fatalf("want ErrActiveBookingExists, got %v", err)
	}

	out, err := br.CancelBookingForUser(ctx, user.ID, b1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "cancelled" {
		t.Fatalf("status %q", out.Status)
	}

	// idempotent cancel
	out2, err := br.CancelBookingForUser(ctx, user.ID, b1.ID)
	if err != nil || out2.Status != "cancelled" {
		t.Fatalf("second cancel: %+v %v", out2, err)
	}
}

func TestBookingRepository_Integration_CreateSlotMissing(t *testing.T) {
	db, cleanup := integrationtest.OpenPostgresAndMigrate(t)
	t.Cleanup(cleanup)
	ctx := context.Background()
	ur := usersrepo.NewUserRepository(db)
	user := &usersmodel.User{ID: uuid.New(), Email: "book-miss@x", PasswordHash: "h", Role: "user"}
	if err := ur.Create(ctx, user); err != nil {
		t.Fatal(err)
	}
	br := repository.NewBookingRepository(db)
	missingSlot := uuid.New()
	b := &model.Booking{ID: uuid.New(), SlotID: missingSlot, UserID: user.ID, Status: "active"}
	err := br.CreateActiveBookingWithSlotLock(ctx, b, missingSlot, time.Now().UTC())
	if !errors.Is(err, repository.ErrBookingSlotNotFound) {
		t.Fatalf("got %v", err)
	}
}
