package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`           // id: UUID, PK
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex"`                   // email: уникальный
	PasswordHash string    `gorm:"type:varchar(255);not null"`                               // hash пароля
	Role         string    `gorm:"type:varchar(10);not null;check:role IN ('admin','user')"` // роль: 'admin' или 'user'
	CreatedAt    time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`     // дата создания
	UpdatedAt    time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`     // дата обновления
}
