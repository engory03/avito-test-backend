package mapper

import (
	"testing"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"github.com/lib/pq"
)

func TestScheduleToJSON_Nil(t *testing.T) {
	if ScheduleToJSON(nil) != nil {
		t.Fatal()
	}
}

func TestScheduleToJSON_OK(t *testing.T) {
	s := &model.Schedule{
		ID: uuid.New(), RoomID: uuid.New(),
		DaysOfWeek: pq.Int64Array{1, 7},
		StartTime:  "09:00", EndTime: "10:00",
	}
	j := ScheduleToJSON(s)
	if j == nil || len(j.DaysOfWeek) != 2 || j.DaysOfWeek[1] != 7 {
		t.Fatalf("%+v", j)
	}
}
