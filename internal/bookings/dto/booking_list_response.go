package dto

type PaginationJSON struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

type ListBookingsAdminResponse struct {
	Bookings   []*BookingJSON `json:"bookings"`
	Pagination PaginationJSON `json:"pagination"`
}
