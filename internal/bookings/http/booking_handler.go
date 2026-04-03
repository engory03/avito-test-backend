package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/dto"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/mapper"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/service"
	_ "github.com/internships-backend/test-backend-engory03/internal/shared/dto"
	httphelper "github.com/internships-backend/test-backend-engory03/internal/shared/http"
)

type BookingHandler struct {
	service service.BookingService
}

func NewBookingHandler(s service.BookingService) *BookingHandler {
	return &BookingHandler{service: s}
}

// MyBookings godoc
// @Summary      Список броней текущего пользователя (только user)
// @Description  Доступно только роли user. Возвращает брони пользователя из JWT. Только брони на будущие слоты (start слота >= now, UTC); прошедшие слоты не включаются.
// @Tags         Bookings
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} dto.ListMyBookingsResponse "Список броней текущего пользователя"
// @Failure      401 {object} shareddto.APIErrorResponse "Не авторизован"
// @Failure      403 {object} shareddto.APIErrorResponse "Доступ запрещён (только user)"
// @Failure      500 {object} shareddto.APIErrorResponse "Внутренняя ошибка сервера"
// @Router       /bookings/my [get]
func (h *BookingHandler) MyBookings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "user_id not found in context")
		return
	}
	id, ok := userID.(uuid.UUID)
	if !ok {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid user_id type")
		return
	}

	list, err := h.service.ListMyFutureBookings(id)
	if err != nil {
		httphelper.JSONAPIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	out := make([]*dto.BookingJSON, 0, len(list))
	for _, b := range list {
		out = append(out, mapper.BookingToJSON(b))
	}
	c.JSON(http.StatusOK, dto.ListMyBookingsResponse{Bookings: out})
}
