package http

import (
	"github.com/gin-gonic/gin"
	"github.com/internships-backend/test-backend-engory03/internal/auth/middleware"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/service"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
)

func RegisterBookingRoutes(r *gin.RouterGroup, bookingService service.BookingService, jwtManager *utils.JWTManager) {
	handler := NewBookingHandler(bookingService)

	g := r.Group("/bookings")
	g.Use(middleware.AuthMiddleware(jwtManager), middleware.RoleMiddleware("user"))
	{
		g.GET("/my", handler.MyBookings)
	}
}
