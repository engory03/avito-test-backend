package mapper

import (
	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
)

func RoomToJSON(r *model.Room) *dto.RoomJSON {
	if r == nil {
		return nil
	}
	created := r.CreatedAt
	return &dto.RoomJSON{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Capacity:    r.Capacity,
		CreatedAt:   &created,
	}
}
