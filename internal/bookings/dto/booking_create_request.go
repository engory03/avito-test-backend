package dto

import "github.com/google/uuid"

// CreateBookingRequest — запрос на создание бронирования слота.
// @Description Запрос на создание бронирования
// @Name CreateBookingRequest
type CreateBookingRequest struct {
	SlotID               uuid.UUID `json:"slotId" binding:"required" example:"a4f8e981-df5d-4cc4-ab2f-520e6cfd35c1"`
	CreateConferenceLink bool      `json:"createConferenceLink" example:"true"`
}
