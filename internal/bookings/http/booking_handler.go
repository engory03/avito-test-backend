package http

import (
	"errors"
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

// CreateBooking godoc
// @Summary      Создать бронь на слот (только user). Опционально — запросить ссылку на конференцию.
// @Description  Доступно только роли user. `userId` из JWT, не из тела. При `start` слота < now (UTC) — 400. Опционально `createConferenceLink` (по умолчанию false) — мок Conference Service.
// @Tags         Bookings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateBookingRequest true "slotId обязателен; createConferenceLink опционально"
// @Success      201 {object} dto.CreateBookingResponse "Бронь создана"
// @Failure      400 {object} shareddto.APIErrorResponse "Неверный запрос (в т.ч. слот в прошлом)"
// @Failure      401 {object} shareddto.APIErrorResponse "Не авторизован"
// @Failure      403 {object} shareddto.APIErrorResponse "Доступ запрещён (только user)"
// @Failure      404 {object} shareddto.APIErrorResponse "Слот не найден"
// @Failure      409 {object} shareddto.APIErrorResponse "Слот уже занят"
// @Failure      500 {object} shareddto.APIErrorResponse "Внутренняя ошибка сервера"
// @Router       /bookings/create [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
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

	var req dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	b, err := h.service.CreateBooking(c.Request.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlotNotFound):
			httphelper.JSONAPIError(c, http.StatusNotFound, "SLOT_NOT_FOUND", err.Error())
		case errors.Is(err, service.ErrSlotInPast):
			httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		case errors.Is(err, service.ErrSlotAlreadyBooked):
			httphelper.JSONAPIError(c, http.StatusConflict, "SLOT_ALREADY_BOOKED", "slot is already booked")
		case errors.Is(err, service.ErrConferenceFailed):
			httphelper.JSONAPIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "conference service error")
		default:
			httphelper.JSONAPIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	c.JSON(http.StatusCreated, dto.CreateBookingResponse{Booking: mapper.BookingToJSON(b)})
}

// CancelBooking godoc
// @Summary      Отменить бронь (только своя бронь, только user)
// @Description  Идемпотентно: повтор на уже отменённой брони — 200 с `status: cancelled`. Чужая бронь — 403 с сообщением из спецификации. Нет брони — 404.
// @Tags         Bookings
// @Security     BearerAuth
// @Produce      json
// @Param        bookingId path string true "Идентификатор брони (UUID)"
// @Success      200 {object} dto.CancelBookingResponse "Бронь отменена или уже была отменена"
// @Failure      401 {object} shareddto.APIErrorResponse "Не авторизован"
// @Failure      403 {object} shareddto.APIErrorResponse "Не своя бронь или роль не user"
// @Failure      404 {object} shareddto.APIErrorResponse "Бронь не найдена"
// @Failure      500 {object} shareddto.APIErrorResponse "Внутренняя ошибка сервера"
// @Router       /bookings/{bookingId}/cancel [post]
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "user_id not found in context")
		return
	}
	uid, ok := userID.(uuid.UUID)
	if !ok {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid user_id type")
		return
	}

	bookingID, err := uuid.Parse(c.Param("bookingId"))
	if err != nil {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid bookingId in path")
		return
	}

	b, err := h.service.CancelBooking(uid, bookingID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			httphelper.JSONAPIError(c, http.StatusNotFound, "BOOKING_NOT_FOUND", err.Error())
		case errors.Is(err, service.ErrCancelNotOwner):
			httphelper.JSONAPIError(c, http.StatusForbidden, "FORBIDDEN", "cannot cancel another user's booking")
		default:
			httphelper.JSONAPIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	c.JSON(http.StatusOK, dto.CancelBookingResponse{Booking: mapper.BookingToJSON(b)})
}
