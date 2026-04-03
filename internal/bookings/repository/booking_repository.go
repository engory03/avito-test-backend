package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/model"
	roomsModel "github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"gorm.io/gorm"
)

// ErrActiveBookingExists — в транзакции: на слот уже есть активная бронь.
var ErrActiveBookingExists = errors.New("active booking exists for slot")

type BookingRepository interface {
	ListByUserWithFutureSlots(userID uuid.UUID, nowUTC time.Time) ([]*model.Booking, error)
	GetSlotByID(slotID uuid.UUID) (*roomsModel.Slot, error)
	CreateActiveBookingIfSlotFree(booking *model.Booking, slotID uuid.UUID) error
	GetByID(bookingID uuid.UUID) (*model.Booking, error)
	MarkActiveBookingCancelled(bookingID uuid.UUID) error
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) ListByUserWithFutureSlots(userID uuid.UUID, nowUTC time.Time) ([]*model.Booking, error) {
	var rows []*model.Booking
	err := r.db.Model(&model.Booking{}).
		Joins("INNER JOIN slots ON slots.id = bookings.slot_id").
		Where("bookings.user_id = ? AND slots.start_at >= ?", userID, nowUTC).
		Order("slots.start_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *bookingRepository) GetSlotByID(slotID uuid.UUID) (*roomsModel.Slot, error) {
	var s roomsModel.Slot
	err := r.db.First(&s, "id = ?", slotID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *bookingRepository) CreateActiveBookingIfSlotFree(booking *model.Booking, slotID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var n int64
		if err := tx.Model(&model.Booking{}).
			Where("slot_id = ? AND status = ?", slotID, "active").
			Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return ErrActiveBookingExists
		}
		return tx.Create(booking).Error
	})
}

func (r *bookingRepository) GetByID(bookingID uuid.UUID) (*model.Booking, error) {
	var b model.Booking
	err := r.db.First(&b, "id = ?", bookingID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *bookingRepository) MarkActiveBookingCancelled(bookingID uuid.UUID) error {
	return r.db.Model(&model.Booking{}).
		Where("id = ? AND status = ?", bookingID, "active").
		Updates(map[string]interface{}{
			"status":     "cancelled",
			"updated_at": time.Now().UTC(),
		}).Error
}
