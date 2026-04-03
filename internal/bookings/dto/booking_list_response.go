package dto

// PaginationJSON — метаданные постраничной выдачи.
// @Description Параметры пагинации
// @Name PaginationJSON
type PaginationJSON struct {
	Page     int `json:"page" example:"1"`
	PageSize int `json:"pageSize" example:"20"`
	Total    int `json:"total" example:"42"`
}

// ListBookingsAdminResponse — список бронирований для администратора с пагинацией.
// @Description Список бронирований (админ) с пагинацией
// @Name ListBookingsAdminResponse
type ListBookingsAdminResponse struct {
	Bookings   []*BookingJSON `json:"bookings"`
	Pagination PaginationJSON `json:"pagination"`
}
