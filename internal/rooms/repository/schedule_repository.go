package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"gorm.io/gorm"
)

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *model.Schedule) error
	ExistsByRoomID(ctx context.Context, roomID uuid.UUID) (bool, error)
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error)
}

type scheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) Create(ctx context.Context, schedule *model.Schedule) error {
	return r.db.WithContext(ctx).Create(schedule).Error
}

func (r *scheduleRepository) ExistsByRoomID(ctx context.Context, roomID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Schedule{}).Where("room_id = ?", roomID).Count(&count).Error
	return count > 0, err
}

func (r *scheduleRepository) GetByRoomID(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error) {
	var s model.Schedule
	err := r.db.WithContext(ctx).Where("room_id = ?", roomID).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
