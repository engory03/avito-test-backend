package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/model"
	"gorm.io/gorm"
)

type BookingRepository interface {
	ListByUserWithFutureSlots(userID uuid.UUID, nowUTC time.Time) ([]*model.Booking, error)
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
