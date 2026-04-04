package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/mapper"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/service"
	_ "github.com/internships-backend/test-backend-engory03/internal/shared/dto"
	httphelper "github.com/internships-backend/test-backend-engory03/internal/shared/http"
)

type SlotHandler struct {
	service service.SlotService
}

func NewSlotHandler(s service.SlotService) *SlotHandler {
	return &SlotHandler{service: s}
}

// ListSlots godoc
// @Summary      Список доступных для бронирования слотов по переговорке и дате (admin и user)
// @Description  Возвращает слоты, не занятые активной бронью, для указанной переговорки на указанную дату. Все даты и время в UTC. Параметр date обязателен (ISO 8601 date, например 2024-06-10). Без расписания или вне рабочих дней — пустой массив slots.
// @Tags         Slots
// @Security     BearerAuth
// @Produce      json
// @Param        roomId path string true "Идентификатор переговорки (UUID)"
// @Param        date query string true "Дата в формате ISO 8601 (YYYY-MM-DD), UTC-календарный день" example(2024-06-10)
// @Success      200 {object} dto.ListSlotsResponse "Список доступных слотов"
// @Failure      400 {object} shareddto.APIErrorResponse "Неверный запрос (отсутствует или некорректен параметр date)"
// @Failure      401 {object} shareddto.APIErrorResponse "Не авторизован"
// @Failure      404 {object} shareddto.APIErrorResponse "Переговорка не найдена"
// @Failure      500 {object} shareddto.APIErrorResponse "Внутренняя ошибка сервера"
// @Router       /rooms/{roomId}/slots/list [get]
func (h *SlotHandler) ListSlots(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("roomId"))
	if err != nil {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId in path")
		return
	}

	var q dto.SlotListQuery
	if bindErr := c.ShouldBindQuery(&q); bindErr != nil {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", bindErr.Error())
		return
	}

	slots, err := h.service.ListAvailableSlots(roomID, q.Date)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidSlotListDate):
			httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		case errors.Is(err, service.ErrRoomNotFound):
			httphelper.JSONAPIError(c, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error())
		default:
			httphelper.JSONAPIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	c.JSON(http.StatusOK, dto.ListSlotsResponse{Slots: mapper.SlotsToJSON(slots)})
}
