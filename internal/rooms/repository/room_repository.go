package repository

import (
	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"gorm.io/gorm"
)

type RoomRepository interface {
	Create(room *model.Room) error
	ListAll() ([]*model.Room, error)
	ExistsByID(id uuid.UUID) (bool, error)
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

func (r *roomRepository) ListAll() ([]*model.Room, error) {
	var rooms []*model.Room
	if err := r.db.Order("created_at DESC").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *roomRepository) ExistsByID(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Room{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
