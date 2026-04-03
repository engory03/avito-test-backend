package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"gorm.io/gorm"
)

type ScheduleRepository interface {
	Create(schedule *model.Schedule) error
	ExistsByRoomID(roomID uuid.UUID) (bool, error)
	GetByRoomID(roomID uuid.UUID) (*model.Schedule, error)
}

type scheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) Create(schedule *model.Schedule) error {
	return r.db.Create(schedule).Error
}

func (r *scheduleRepository) ExistsByRoomID(roomID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Schedule{}).Where("room_id = ?", roomID).Count(&count).Error
	return count > 0, err
}

func (r *scheduleRepository) GetByRoomID(roomID uuid.UUID) (*model.Schedule, error) {
	var s model.Schedule
	err := r.db.Where("room_id = ?", roomID).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
