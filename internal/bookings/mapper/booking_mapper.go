package mapper

import (
	"github.com/internships-backend/test-backend-engory03/internal/bookings/dto"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/model"
)

func BookingToJSON(b *model.Booking) *dto.BookingJSON {
	if b == nil {
		return nil
	}
	created := b.CreatedAt
	return &dto.BookingJSON{
		ID:             b.ID,
		SlotID:         b.SlotID,
		UserID:         b.UserID,
		Status:         b.Status,
		ConferenceLink: b.ConferenceLink,
		CreatedAt:      &created,
	}
}
