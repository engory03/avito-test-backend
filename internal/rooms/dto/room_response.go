package dto

import (
	"time"

	"github.com/google/uuid"
)

// RoomJSON — данные переговорной в JSON-представлении.
// @Description Данные переговорной
// @Name RoomJSON
type RoomJSON struct {
	ID          uuid.UUID  `json:"id" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`                   // Идентификатор переговорной
	Name        string     `json:"name" example:"Большая переговорная"`                                 // Название переговорной
	Description *string    `json:"description" example:"Просторная переговорная с проектором и доской"` // Описание переговорной
	Capacity    *int       `json:"capacity" example:"20"`                                               // Вместимость
	CreatedAt   *time.Time `json:"createdAt" example:"2024-04-12T09:41:00Z"`                            // Время создания (ISO8601)
}

// CreateRoomResponse — ответ после создания переговорной.
// @Description Ответ создания переговорной
// @Name CreateRoomResponse
type CreateRoomResponse struct {
	Room *RoomJSON `json:"room"`
}

// ListRoomsResponse — список переговорных в JSON-представлении.
// @Description Список переговорных
// @Name ListRoomsResponse
type ListRoomsResponse struct {
	Rooms []*RoomJSON `json:"rooms"`
}
