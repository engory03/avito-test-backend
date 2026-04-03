package dto

type SlotJSON struct {
	ID     string `json:"id" example:"a8098c1a-f86e-11da-bd1a-00112444be1e"`
	RoomID string `json:"roomId" example:"a8098c1a-f86e-11da-bd1a-00112444be1e"`
	Start  string `json:"start" example:"2024-06-10T09:00:00Z"`
	End    string `json:"end" example:"2024-06-10T09:30:00Z"`
}

type ListSlotsResponse struct {
	Slots []*SlotJSON `json:"slots"`
}
