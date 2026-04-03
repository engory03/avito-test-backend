package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/conference"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/dto"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/model"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/repository"
)

var (
	ErrSlotNotFound      = errors.New("slot not found")
	ErrSlotInPast        = errors.New("slot is in the past")
	ErrSlotAlreadyBooked = errors.New("slot is already booked")
	ErrConferenceFailed  = errors.New("conference link could not be created")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrCancelNotOwner    = errors.New("cannot cancel another user's booking")
)

type BookingService interface {
	ListMyFutureBookings(userID uuid.UUID) ([]*model.Booking, error)
	CreateBooking(ctx context.Context, userID uuid.UUID, req dto.CreateBookingRequest) (*model.Booking, error)
	CancelBooking(userID, bookingID uuid.UUID) (*model.Booking, error)
}

type bookingService struct {
	repo       repository.BookingRepository
	conference conference.Client
}

func NewBookingService(r repository.BookingRepository, c conference.Client) BookingService {
	return &bookingService{repo: r, conference: c}
}

func (s *bookingService) ListMyFutureBookings(userID uuid.UUID) ([]*model.Booking, error) {
	return s.repo.ListByUserWithFutureSlots(userID, time.Now().UTC())
}

func (s *bookingService) CreateBooking(ctx context.Context, userID uuid.UUID, req dto.CreateBookingRequest) (*model.Booking, error) {
	now := time.Now().UTC()

	slot, err := s.repo.GetSlotByID(req.SlotID)
	if err != nil {
		return nil, err
	}
	if slot == nil {
		return nil, ErrSlotNotFound
	}
	if slot.StartAt.UTC().Before(now) {
		return nil, ErrSlotInPast
	}

	var confLink *string
	if req.CreateConferenceLink {
		url, err := s.conference.CreateLink(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrConferenceFailed, err)
		}
		confLink = &url
	}

	b := &model.Booking{
		ID:             uuid.New(),
		SlotID:         req.SlotID,
		UserID:         userID,
		Status:         "active",
		ConferenceLink: confLink,
	}

	if err := s.repo.CreateActiveBookingIfSlotFree(b, req.SlotID); err != nil {
		if errors.Is(err, repository.ErrActiveBookingExists) {
			return nil, ErrSlotAlreadyBooked
		}
		return nil, err
	}
	return b, nil
}

func (s *bookingService) CancelBooking(userID, bookingID uuid.UUID) (*model.Booking, error) {
	b, err := s.repo.GetByID(bookingID)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, ErrBookingNotFound
	}
	if b.UserID != userID {
		return nil, ErrCancelNotOwner
	}
	if b.Status == "cancelled" {
		return b, nil
	}
	if b.Status != "active" {
		return nil, fmt.Errorf("unexpected booking status: %s", b.Status)
	}
	if err := s.repo.MarkActiveBookingCancelled(bookingID); err != nil {
		return nil, err
	}
	updated, err := s.repo.GetByID(bookingID)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrBookingNotFound
	}
	return updated, nil
}
