package http

import (
	"github.com/gin-gonic/gin"
	"github.com/internships-backend/test-backend-engory03/internal/auth/service"
)

func RegisterAuthRoutes(r *gin.RouterGroup, authService service.AuthService) {
	authHandler := NewAuthHandler(authService)

	authGroup := r.Group("")
	authGroup.POST("/dummyLogin", authHandler.DummyLogin)
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.POST("/logout", authHandler.Logout)
}
