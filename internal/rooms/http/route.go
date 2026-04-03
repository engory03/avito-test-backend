package http

import (
	"github.com/gin-gonic/gin"
	"github.com/internships-backend/test-backend-engory03/internal/auth/middleware"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/service"
)

func RegisterRoomRoutes(r *gin.RouterGroup, roomService service.RoomService, jwtManager *utils.JWTManager) {
	handler := NewRoomHandler(roomService)

	g := r.Group("/rooms")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("/list", handler.ListRooms)

		admin := g.Group("")
		admin.Use(middleware.RoleMiddleware("admin"))
		{
			admin.POST("/create", handler.CreateRoom)
		}
	}
}
