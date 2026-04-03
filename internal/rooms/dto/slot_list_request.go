package dto

// SlotListQuery — query-параметры GET .../slots/list.
// @Description Параметры списка слотов по дате
// @Name SlotListQuery
type SlotListQuery struct {
	Date string `form:"date" binding:"required" example:"2024-06-10"`
}
