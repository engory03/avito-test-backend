package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

var (
	ErrScheduleRoomNotFound     = errors.New("room not found for schedule")
	ErrScheduleDuplicateForRoom = errors.New("schedule already exists for room")
)

type ScheduleRepository interface {
	CreateIfRoomExists(ctx context.Context, schedule *model.Schedule) error
	ExistsByRoomID(ctx context.Context, roomID uuid.UUID) (bool, error)
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error)
}

type scheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

func isPostgresUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

// CreateIfRoomExists: READ COMMITTED — проверка комнаты и INSERT расписания в одной транзакции.
func (r *scheduleRepository) CreateIfRoomExists(ctx context.Context, schedule *model.Schedule) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var n int64
		if err := tx.Model(&model.Room{}).Where("id = ?", schedule.RoomID).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			return ErrScheduleRoomNotFound
		}
		if err := tx.Create(schedule).Error; err != nil {
			if isPostgresUniqueViolation(err) {
				return ErrScheduleDuplicateForRoom
			}
			return err
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
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
