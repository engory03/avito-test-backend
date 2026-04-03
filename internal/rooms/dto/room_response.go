package dto

import (
	"time"

	"github.com/google/uuid"
)

type RoomJSON struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Capacity    *int       `json:"capacity"`
	CreatedAt   *time.Time `json:"createdAt"`
}

type CreateRoomResponse struct {
	Room *RoomJSON `json:"room"`
}

type ListRoomsResponse struct {
	Rooms []*RoomJSON `json:"rooms"`
}
