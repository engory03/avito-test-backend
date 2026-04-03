package http

import (
	"github.com/gin-gonic/gin"
	"github.com/internships-backend/test-backend-engory03/internal/auth/middleware"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/service"
)

func RegisterRoomRoutes(
	r *gin.RouterGroup,
	roomService service.RoomService,
	scheduleService service.ScheduleService,
	slotService service.SlotService,
	jwtManager *utils.JWTManager,
) {
	roomHandler := NewRoomHandler(roomService)
	scheduleHandler := NewScheduleHandler(scheduleService)
	slotHandler := NewSlotHandler(slotService)

	g := r.Group("/rooms")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("/list", roomHandler.ListRooms)
		g.GET("/:roomId/slots/list", slotHandler.ListSlots)

		admin := g.Group("")
		admin.Use(middleware.RoleMiddleware("admin"))
		{
			admin.POST("/create", roomHandler.CreateRoom)
			admin.POST("/:roomId/schedule/create", scheduleHandler.CreateSchedule)
		}
	}
}
