package repository

import (
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"gorm.io/gorm"
)

type RoomRepository interface {
	Create(room *model.Room) error
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(room *model.Room) error {
	return r.db.Create(room).Error
}
