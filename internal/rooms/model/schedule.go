package model

import (
	"time"

	"github.com/google/uuid"
)

type Schedule struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`       // id: UUID, PK
	RoomID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`                       // room_id: UUID, уникально, FK -> rooms(id)
	DaysOfWeek []int     `gorm:"type:integer[];not null"`                              // days_of_week: массив рабочих дней (1=Пн, 7=Вс)
	StartTime  string    `gorm:"type:varchar(5);not null"`                             // start_time: начало окна (HH:MM, UTC)
	EndTime    string    `gorm:"type:varchar(5);not null"`                             // end_time: конец окна (HH:MM, UTC)
	CreatedAt  time.Time `gorm:"type:timestamp with time zone;not null;default:now()"` // дата создания
}
