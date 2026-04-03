package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/model"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/repository"
)

type BookingService interface {
	ListMyFutureBookings(userID uuid.UUID) ([]*model.Booking, error)
}

type bookingService struct {
	repo repository.BookingRepository
}

func NewBookingService(r repository.BookingRepository) BookingService {
	return &bookingService{repo: r}
}

func (s *bookingService) ListMyFutureBookings(userID uuid.UUID) ([]*model.Booking, error) {
	return s.repo.ListByUserWithFutureSlots(userID, time.Now().UTC())
}
