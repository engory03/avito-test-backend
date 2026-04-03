package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/mapper"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/service"
	httphelper "github.com/internships-backend/test-backend-engory03/internal/shared/http"
)

type RoomHandler struct {
	service service.RoomService
}

func NewRoomHandler(s service.RoomService) *RoomHandler {
	return &RoomHandler{service: s}
}

// CreateRoom godoc
// @Summary      Создать переговорку (только admin)
// @Tags         Rooms
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        data body dto.CreateRoomRequest true "Данные для создания переговорки"
// @Success      201 {object} dto.CreateRoomResponse "Переговорка создана"
// @Failure      400 {object} httphelper.ErrorResponse "Неверный запрос"
// @Failure      401 {object} httphelper.ErrorResponse "Не авторизован"
// @Failure      403 {object} httphelper.ErrorResponse "Доступ запрещён (требуется роль admin)"
// @Failure      500 {object} httphelper.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router       /rooms/create [post]
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	room, err := h.service.CreateRoom(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCapacity) {
			httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		httphelper.JSONAPIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	c.JSON(http.StatusCreated, dto.CreateRoomResponse{Room: mapper.RoomToJSON(room)})
}

// ListRooms godoc
// @Summary      Список переговорок (admin и user)
// @Tags         Rooms
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} dto.ListRoomsResponse "Список переговорок"
// @Failure      401 {object} httphelper.ErrorResponse "Не авторизован"
// @Failure      500 {object} httphelper.InternalErrorResponse "Внутренняя ошибка сервера"
// @Router       /rooms/list [get]
func (h *RoomHandler) ListRooms(c *gin.Context) {
	rooms, err := h.service.ListRooms()
	if err != nil {
		httphelper.JSONAPIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	out := make([]*dto.RoomJSON, 0, len(rooms))
	for _, r := range rooms {
		out = append(out, mapper.RoomToJSON(r))
	}
	c.JSON(http.StatusOK, dto.ListRoomsResponse{Rooms: out})
}
