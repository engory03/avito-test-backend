package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	bookingsModel "github.com/internships-backend/test-backend-engory03/internal/bookings/model"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SlotRepository interface {
	UpsertGeneratedSlotsBatch(ctx context.Context, slots []model.Slot) error
	ListActiveBookedSlotIDs(ctx context.Context, slotIDs []uuid.UUID) (map[uuid.UUID]struct{}, error)
}

type slotRepository struct {
	db *gorm.DB
}

func NewSlotRepository(db *gorm.DB) SlotRepository {
	return &slotRepository{db: db}
}

// UpsertGeneratedSlotsBatch: READ COMMITTED — все вставки слотов атомарно.
func (r *slotRepository) UpsertGeneratedSlotsBatch(ctx context.Context, slots []model.Slot) error {
	if len(slots) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range slots {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "room_id"}, {Name: "start_at"}},
				DoNothing: true,
			}).Create(&slots[i]).Error; err != nil {
				return err
			}
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
}

func (r *slotRepository) ListActiveBookedSlotIDs(ctx context.Context, slotIDs []uuid.UUID) (map[uuid.UUID]struct{}, error) {
	if len(slotIDs) == 0 {
		return map[uuid.UUID]struct{}{}, nil
	}
	var rows []uuid.UUID
	err := r.db.WithContext(ctx).Model(&bookingsModel.Booking{}).
		Where("slot_id IN ? AND status = ?", slotIDs, "active").
		Distinct("slot_id").
		Pluck("slot_id", &rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]struct{}, len(rows))
	for _, id := range rows {
		out[id] = struct{}{}
	}
	return out, nil
}
