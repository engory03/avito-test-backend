package mapper

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
)

func TestSlotToJSON_Nil(t *testing.T) {
	if SlotToJSON(nil) != nil {
		t.Fatal()
	}
}

func TestSlotToJSON_AndSlotsToJSON(t *testing.T) {
	id := uuid.New()
	rid := uuid.New()
	start := time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	s := model.Slot{ID: id, RoomID: rid, StartAt: start, EndAt: end}
	j := SlotToJSON(&s)
	if j.ID != id.String() || j.RoomID != rid.String() {
		t.Fatalf("%+v", j)
	}
	list := SlotsToJSON([]model.Slot{s})
	if len(list) != 1 || list[0].ID != id.String() {
		t.Fatal(list)
	}
}
