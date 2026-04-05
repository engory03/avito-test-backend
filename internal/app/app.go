package app

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	authHttp "github.com/internships-backend/test-backend-engory03/internal/auth/http"
	authRepo "github.com/internships-backend/test-backend-engory03/internal/auth/repository"
	authService "github.com/internships-backend/test-backend-engory03/internal/auth/service"

	bookingsConference "github.com/internships-backend/test-backend-engory03/internal/bookings/conference"
	bookingsHttp "github.com/internships-backend/test-backend-engory03/internal/bookings/http"
	bookingsRepo "github.com/internships-backend/test-backend-engory03/internal/bookings/repository"
	bookingsService "github.com/internships-backend/test-backend-engory03/internal/bookings/service"

	roomsHttp "github.com/internships-backend/test-backend-engory03/internal/rooms/http"
	roomsRepo "github.com/internships-backend/test-backend-engory03/internal/rooms/repository"
	roomsService "github.com/internships-backend/test-backend-engory03/internal/rooms/service"

	usersHttp "github.com/internships-backend/test-backend-engory03/internal/users/http"
	usersRepo "github.com/internships-backend/test-backend-engory03/internal/users/repository"
	usersService "github.com/internships-backend/test-backend-engory03/internal/users/service"

	"github.com/internships-backend/test-backend-engory03/internal/db"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
	"gorm.io/gorm"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/internships-backend/test-backend-engory03/docs"
)

// EngineConfig drives JWT settings for the HTTP engine (same semantics as cmd/app env).
type EngineConfig struct {
	JWTAccessSecret   string
	JWTRefreshSecret  string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

// MigrateDB applies SQL migrations (golang-migrate) embedded in internal/db.
func MigrateDB(gormDB *gorm.DB) error {
	return db.RunMigrations(gormDB)
}

// NewEngine builds the Gin engine with API routes (shared by cmd/app and e2e).
func NewEngine(db *gorm.DB, cfg EngineConfig) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))

	jwtManager := utils.NewJWTManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret)
	jwtManager.AccessTokenTTL = cfg.AccessExpiration
	jwtManager.RefreshTokenTTL = cfg.RefreshExpiration

	userRepo := usersRepo.NewUserRepository(db)
	userService := usersService.NewUserService(userRepo)

	refreshTokenRepo := authRepo.NewRefreshTokenRepository(db)
	authService := authService.NewAuthService(userRepo, refreshTokenRepo, jwtManager)

	roomRepo := roomsRepo.NewRoomRepository(db)
	roomService := roomsService.NewRoomService(roomRepo)

	scheduleRepo := roomsRepo.NewScheduleRepository(db)
	scheduleService := roomsService.NewScheduleService(scheduleRepo)

	slotRepo := roomsRepo.NewSlotRepository(db)
	slotService := roomsService.NewSlotService(roomRepo, scheduleRepo, slotRepo)

	bookingRepo := bookingsRepo.NewBookingRepository(db)
	conferenceClient := bookingsConference.NewMockClientFromEnv()
	bookingService := bookingsService.NewBookingService(bookingRepo, conferenceClient)

	api := r.Group("/api")
	authHttp.RegisterAuthRoutes(api, authService)
	usersHttp.RegisterUserRoutes(api, userService, jwtManager)
	bookingsHttp.RegisterBookingRoutes(api, bookingService, jwtManager)
	roomsHttp.RegisterRoomRoutes(api, roomService, scheduleService, slotService, jwtManager)

	r.GET("/_info", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.GET("/docs/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.PersistAuthorization(true),
	))

	return r
}
