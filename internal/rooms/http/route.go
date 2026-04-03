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
	jwtManager *utils.JWTManager,
) {
	roomHandler := NewRoomHandler(roomService)
	scheduleHandler := NewScheduleHandler(scheduleService)

	g := r.Group("/rooms")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("/list", roomHandler.ListRooms)

		admin := g.Group("")
		admin.Use(middleware.RoleMiddleware("admin"))
		{
			admin.POST("/create", roomHandler.CreateRoom)
			admin.POST("/:roomId/schedule/create", scheduleHandler.CreateSchedule)
		}
	}
}
