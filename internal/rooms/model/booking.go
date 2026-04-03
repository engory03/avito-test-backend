package model

import (
	"time"

	"github.com/google/uuid"
)

// Booking represents a booking for a slot. It is used to exclude slots with active bookings.
type Booking struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SlotID         uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID         uuid.UUID `gorm:"type:uuid;not null"`
	Status         string    `gorm:"type:varchar(10);not null;default:active;check:status IN ('active','cancelled')"`
	ConferenceLink *string   `gorm:"type:text"`
	CreatedAt      time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
	UpdatedAt      time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
}
