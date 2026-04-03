package model

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`       // id: UUID, PK
	Name        string    `gorm:"type:varchar(255);not null"`                           // name: название переговорки
	Description *string   `gorm:"type:text"`                                            // description: описание (опционально)
	Capacity    *int      `gorm:"type:integer;check:capacity > 0"`                      // capacity: вместимость (опционально)
	CreatedAt   time.Time `gorm:"type:timestamp with time zone;not null;default:now()"` // дата создания
	UpdatedAt   time.Time `gorm:"type:timestamp with time zone;not null;default:now()"` // дата обновления
}
