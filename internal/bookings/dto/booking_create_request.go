package dto

import "github.com/google/uuid"

type CreateBookingRequest struct {
	SlotID               uuid.UUID `json:"slotId" binding:"required"`
	CreateConferenceLink bool      `json:"createConferenceLink"`
}
