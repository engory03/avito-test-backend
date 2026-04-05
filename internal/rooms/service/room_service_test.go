package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
)

type fakeRoomRepo struct {
	create func(context.Context, *model.Room) error
	list   func(context.Context) ([]*model.Room, error)
}

func (f *fakeRoomRepo) Create(ctx context.Context, room *model.Room) error {
	if f.create != nil {
		return f.create(ctx, room)
	}
	return nil
}

func (f *fakeRoomRepo) ListAll(ctx context.Context) ([]*model.Room, error) {
	if f.list != nil {
		return f.list(ctx)
	}
	return nil, nil
}

func (f *fakeRoomRepo) ExistsByID(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}

func TestCreateRoom_InvalidCapacity(t *testing.T) {
	capacity := 0
	svc := NewRoomService(&fakeRoomRepo{})
	_, err := svc.CreateRoom(context.Background(), dto.CreateRoomRequest{Name: "A", Capacity: &capacity})
	if err != ErrInvalidCapacity {
		t.Fatalf("got %v", err)
	}
}

func TestCreateRoom_Success(t *testing.T) {
	desc := "d"
	capacity := 5
	var saved *model.Room
	repo := &fakeRoomRepo{
		create: func(_ context.Context, r *model.Room) error {
			saved = r
			return nil
		},
	}
	svc := NewRoomService(repo)
	got, err := svc.CreateRoom(context.Background(), dto.CreateRoomRequest{Name: "Room", Description: &desc, Capacity: &capacity})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Room" || saved == nil || saved.Name != "Room" {
		t.Fatalf("got %+v saved %+v", got, saved)
	}
}

func TestListRooms(t *testing.T) {
	want := []*model.Room{{ID: uuid.New(), Name: "X"}}
	repo := &fakeRoomRepo{
		list: func(context.Context) ([]*model.Room, error) {
			return want, nil
		},
	}
	svc := NewRoomService(repo)
	got, err := svc.ListRooms(context.Background())
	if err != nil || len(got) != 1 || got[0].Name != "X" {
		t.Fatalf("got %+v err %v", got, err)
	}
}
