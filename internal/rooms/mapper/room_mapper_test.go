package mapper

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
)

func TestRoomToJSON_Nil(t *testing.T) {
	if RoomToJSON(nil) != nil {
		t.Fatal()
	}
}

func TestRoomToJSON_OK(t *testing.T) {
	cap := 4
	r := &model.Room{
		ID: uuid.New(), Name: "N", Capacity: &cap,
		CreatedAt: time.Now().UTC(),
	}
	j := RoomToJSON(r)
	if j == nil || j.Name != "N" || j.Capacity == nil || *j.Capacity != 4 {
		t.Fatalf("%+v", j)
	}
}
