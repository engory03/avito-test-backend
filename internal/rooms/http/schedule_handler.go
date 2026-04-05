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

type ScheduleHandler struct {
	service service.ScheduleService
}

func NewScheduleHandler(s service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: s}
}

// CreateSchedule godoc
// @Summary      Создать расписание для переговорки (только admin)
// @Tags         Schedules
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        roomId path string true "ID переговорки"
// @Param        data body dto.ScheduleCreateRequest true "Данные для создания расписания"
// @Success      201 {object} dto.CreateScheduleResponse "Расписание создано"
// @Failure      400 {object} shareddto.APIErrorResponse "Неверный запрос"
// @Failure      401 {object} shareddto.APIErrorResponse "Не авторизован"
// @Failure      403 {object} shareddto.APIErrorResponse "Требуется роль admin"
// @Failure      404 {object} shareddto.APIErrorResponse "Переговорка не найдена"
// @Failure      409 {object} shareddto.APIErrorResponse "Расписание для этой переговорки уже существует"
// @Failure      500 {object} shareddto.APIErrorResponse "Внутренняя ошибка сервера"
// @Router       /rooms/{roomId}/schedule/create [post]
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("roomId"))
	if err != nil {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId in path")
		return
	}

	var req dto.ScheduleCreateRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", bindErr.Error())
		return
	}

	sched, err := h.service.CreateSchedule(c.Request.Context(), roomID, req)
	if err != nil {
		var valErr *service.ScheduleValidationError
		switch {
		case errors.As(err, &valErr):
			httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", valErr.Message)
		case errors.Is(err, service.ErrRoomNotFound):
			httphelper.JSONAPIError(c, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error())
		case errors.Is(err, service.ErrScheduleExists):
			httphelper.JSONAPIError(c, http.StatusConflict, "SCHEDULE_EXISTS", "schedule for this room already exists and cannot be changed")
		default:
			httphelper.JSONAPIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	c.JSON(http.StatusCreated, dto.CreateScheduleResponse{Schedule: mapper.ScheduleToJSON(sched)})
}
