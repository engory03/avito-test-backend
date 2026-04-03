package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	httphelper "github.com/internships-backend/test-backend-engory03/internal/shared/http"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/mapper"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/service"
)

type RoomHandler struct {
	service service.RoomService
}

func NewRoomHandler(s service.RoomService) *RoomHandler {
	return &RoomHandler{service: s}
}

// CreateRoom handles POST /rooms/create (admin only, Bearer JWT).
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
