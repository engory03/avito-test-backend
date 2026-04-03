package mapper

import (
	"time"

	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
)

func SlotToJSON(s *model.Slot) *dto.SlotJSON {
	if s == nil {
		return nil
	}
	return &dto.SlotJSON{
		ID:     s.ID.String(),
		RoomID: s.RoomID.String(),
		Start:  s.StartAt.UTC().Format(time.RFC3339),
		End:    s.EndAt.UTC().Format(time.RFC3339),
	}
}

func SlotsToJSON(slots []model.Slot) []*dto.SlotJSON {
	out := make([]*dto.SlotJSON, 0, len(slots))
	for i := range slots {
		out = append(out, SlotToJSON(&slots[i]))
	}
	return out
}
