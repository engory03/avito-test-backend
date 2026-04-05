package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	authDummy "github.com/internships-backend/test-backend-engory03/internal/auth/dummy"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"github.com/internships-backend/test-backend-engory03/internal/shared/config"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Fixed IDs keep seed idempotent across runs.
var (
	seedRoomAlpha = uuid.MustParse("11111111-1111-4111-8111-111111111101")
	seedRoomBeta  = uuid.MustParse("11111111-1111-4111-8111-111111111102")
)

func main() {
	config.LoadEnv()

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName, dbPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	if err := authDummy.EnsureUsersInDB(context.Background(), db); err != nil {
		log.Fatalf("ensure dummy users: %v", err)
	}

	cap10 := 10
	descA := "Seeded room for local testing"
	rooms := []model.Room{
		{ID: seedRoomAlpha, Name: "Seed - Conference Alpha", Description: &descA, Capacity: &cap10},
		{ID: seedRoomBeta, Name: "Seed - Conference Beta", Capacity: &cap10},
	}
	for i := range rooms {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rooms[i]).Error; err != nil {
			log.Fatalf("seed room: %v", err)
		}
	}

	schedules := []model.Schedule{
		{
			ID:         uuid.MustParse("22222222-2222-4222-8222-222222222201"),
			RoomID:     seedRoomAlpha,
			DaysOfWeek: pq.Int64Array{1, 2, 3, 4, 5},
			StartTime:  "09:00",
			EndTime:    "17:00",
		},
		{
			ID:         uuid.MustParse("22222222-2222-4222-8222-222222222202"),
			RoomID:     seedRoomBeta,
			DaysOfWeek: pq.Int64Array{1, 2, 3, 4, 5, 6},
			StartTime:  "10:00",
			EndTime:    "14:00",
		},
	}
	for i := range schedules {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&schedules[i]).Error; err != nil {
			log.Fatalf("seed schedule: %v", err)
		}
	}

	log.Println("seed completed (rooms + schedules; slots are generated on first list request)")
}
