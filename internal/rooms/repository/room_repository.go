package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"gorm.io/gorm"
)

type RoomRepository interface {
	Create(ctx context.Context, room *model.Room) error
	ListAll(ctx context.Context) ([]*model.Room, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(ctx context.Context, room *model.Room) error {
	return r.db.WithContext(ctx).Create(room).Error
}

func (r *roomRepository) ListAll(ctx context.Context) ([]*model.Room, error) {
	var rooms []*model.Room
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *roomRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Room{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
