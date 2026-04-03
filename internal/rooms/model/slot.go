package model

import (
	"time"

	"github.com/google/uuid"
)

type Slot struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`                     // id: UUID, PK
	RoomID  uuid.UUID `gorm:"type:uuid;not null;index:idx_room_start,unique"`                     // room_id: UUID, FK → rooms(id), индекс и уникальный составной индекс с start_at
	StartAt time.Time `gorm:"type:timestamp with time zone;not null;index:idx_room_start,unique"` // start_at: начало слота (UTC), составной индекс с room_id, уникальный
	EndAt   time.Time `gorm:"type:timestamp with time zone;not null;check:end_at > start_at"`     // end_at: конец слота (UTC), end_at > start_at
}
