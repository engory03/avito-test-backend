package dto

import (
	"time"

	"github.com/google/uuid"
)

type BookingJSON struct {
	ID             uuid.UUID  `json:"id" example:"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6a"`
	SlotID         uuid.UUID  `json:"slotId" example:"a4f8e981-df5d-4cc4-ab2f-520e6cfd35c1"`
	UserID         uuid.UUID  `json:"userId" example:"3f8b6c1a-84e8-419e-9b12-e97ef70be6ca"`
	Status         string     `json:"status" enums:"active,cancelled" example:"active"`
	ConferenceLink *string    `json:"conferenceLink" example:"https://meet.example.com/meeting123"`
	CreatedAt      *time.Time `json:"createdAt" example:"2024-05-24T15:04:05Z"`
}

type ListMyBookingsResponse struct {
	Bookings []*BookingJSON `json:"bookings"`
}
