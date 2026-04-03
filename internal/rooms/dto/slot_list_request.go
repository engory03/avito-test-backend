package dto

// SlotListQuery binds GET .../slots/list query parameters (api.yaml).
type SlotListQuery struct {
	Date string `form:"date" binding:"required"`
}
