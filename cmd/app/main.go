package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/internships-backend/test-backend-engory03/internal/app"
	"github.com/internships-backend/test-backend-engory03/internal/shared/config"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	if err := app.MigrateDB(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("Database migrated successfully")

	r := app.NewEngine(db, app.EngineConfig{
		JWTAccessSecret:   jwtAccessSecret,
		JWTRefreshSecret:  jwtRefreshSecret,
		AccessExpiration:  accessExp,
		RefreshExpiration: refreshExp,
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Printf("Dev server started at http://localhost:%s\n", port)
		fmt.Printf("Swagger docs at http://localhost:%s/docs/index.html\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("database pool: %v", err)
	} else {
		if cerr := sqlDB.Close(); cerr != nil {
			log.Printf("database close: %v", cerr)
		}
	}
}
