package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	authHttp "github.com/internships-backend/test-backend-engory03/internal/auth/http"
	authModel "github.com/internships-backend/test-backend-engory03/internal/auth/model"
	authRepo "github.com/internships-backend/test-backend-engory03/internal/auth/repository"
	authService "github.com/internships-backend/test-backend-engory03/internal/auth/service"

	roomsHttp "github.com/internships-backend/test-backend-engory03/internal/rooms/http"
	roomsModel "github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	roomsRepo "github.com/internships-backend/test-backend-engory03/internal/rooms/repository"
	roomsService "github.com/internships-backend/test-backend-engory03/internal/rooms/service"

	usersHttp "github.com/internships-backend/test-backend-engory03/internal/users/http"
	usersModel "github.com/internships-backend/test-backend-engory03/internal/users/model"
	usersRepo "github.com/internships-backend/test-backend-engory03/internal/users/repository"
	usersService "github.com/internships-backend/test-backend-engory03/internal/users/service"

	"github.com/internships-backend/test-backend-engory03/internal/shared/config"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/internships-backend/test-backend-engory03/docs"
)

// @title Room Booking Service (Test Task)
// @version 1.0
// @description Сервис бронирования переговорок
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.type http
// @securityDefinitions.scheme bearer
// @securityDefinitions.bearerFormat JWT
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer {your_token}" to authorize

func main() {
	config.LoadEnv()

	// Read environment variables
	port := os.Getenv("PORT")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")

	jwtAccessSecret := os.Getenv("JWT_ACCESS_SECRET")
	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	accessExpStr := os.Getenv("JWT_ACCESS_EXPIRATION")
	refreshExpStr := os.Getenv("JWT_REFRESH_EXPIRATION")

	accessExp, err := utils.ParseDuration(accessExpStr)
	if err != nil {
		log.Fatalf("Invalid JWT_ACCESS_EXPIRATION: %v", err)
	}
	refreshExp, err := utils.ParseDuration(refreshExpStr)
	if err != nil {
		log.Fatalf("Invalid JWT_REFRESH_EXPIRATION: %v", err)
	}

	// Build DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName, dbPort,
	)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Run AutoMigrate
	if err := db.AutoMigrate(
		&usersModel.User{},
		&authModel.RefreshToken{},
		&roomsModel.Room{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	fmt.Println("Database migrated successfully")

	// Initialize router and services
	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))

	jwtManager := utils.NewJWTManager(jwtAccessSecret, jwtRefreshSecret)
	jwtManager.AccessTokenTTL = accessExp
	jwtManager.RefreshTokenTTL = refreshExp

	userRepo := usersRepo.NewUserRepository(db)
	userService := usersService.NewUserService(userRepo)

	roomRepo := roomsRepo.NewRoomRepository(db)
	roomService := roomsService.NewRoomService(roomRepo)

	refreshTokenRepo := authRepo.NewRefreshTokenRepository(db)
	authService := authService.NewAuthService(userRepo, refreshTokenRepo, jwtManager)

	// Run seeds
	// if err := seed.RunSeeds(db); err != nil {
	// 	log.Fatalf("Failed to seed: %v", err)
	// 	panic(err)
	// }

	api := r.Group("/api")
	{
		authHttp.RegisterAuthRoutes(api, authService)
		usersHttp.RegisterUserRoutes(api, userService, jwtManager)
		roomsHttp.RegisterRoomRoutes(api, roomService, jwtManager)
	}

	// Swagger docs
	r.GET("/docs/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.PersistAuthorization(true),
	))

	fmt.Printf("Dev server started at http://localhost:%s\n", port)
	fmt.Printf("Swagger docs at http://localhost:%s/docs/index.html\n", port)

	r.Run(":" + port)
}
