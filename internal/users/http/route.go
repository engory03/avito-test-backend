package http

import (
	"github.com/gin-gonic/gin"
	"github.com/internships-backend/test-backend-engory03/internal/auth/middleware"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
	"github.com/internships-backend/test-backend-engory03/internal/users/service"
)

func RegisterUserRoutes(r *gin.RouterGroup, userService service.UserService, jwtManager *utils.JWTManager) {
	userHandler := NewUserHandler(userService)

	protected := r.Group("/users")
	protected.Use(middleware.AuthMiddleware(jwtManager))
	{
		protected.GET("", userHandler.GetAllUsers)
		protected.GET("/me", userHandler.GetCurrentUser)
		protected.GET("/:id", userHandler.GetUserById)
	}
}
