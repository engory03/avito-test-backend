package mapper

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/model"
)

func TestBookingToJSON_Nil(t *testing.T) {
	if BookingToJSON(nil) != nil {
		t.Fatal()
	}
}

func TestBookingToJSON_OK(t *testing.T) {
	link := "https://x"
	ts := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	b := &model.Booking{
		ID: uuid.New(), SlotID: uuid.New(), UserID: uuid.New(),
		Status: "active", ConferenceLink: &link, CreatedAt: ts,
	}
	j := BookingToJSON(b)
	if j == nil || j.ID != b.ID || j.ConferenceLink == nil || *j.ConferenceLink != link {
		t.Fatalf("%+v", j)
	}
}
